package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	fsym      bool
	ssym      bool
	build     bool
	buildArgs string
	buildEnvs ListOpts
)

type ListOpts []string

func (opts *ListOpts) String() string {
	return fmt.Sprint(*opts)
}

func (opts *ListOpts) Set(value string) error {
	*opts = append(*opts, value)
	return nil
}

func exitPrint(con string) {
	fmt.Fprintln(os.Stderr, con)
	os.Exit(2)
}

type walker interface {
	isExclude(string) bool
	isEmpty(string) bool
	relName(string) string
	virPath(string) string
	compress(string, string, os.FileInfo) (bool, error)
	walkRoot(string) error
}

type byName []os.FileInfo

func (f byName) Len() int           { return len(f) }
func (f byName) Less(i, j int) bool { return f[i].Name() < f[j].Name() }
func (f byName) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }

type walkFileTree struct {
	wak           walker
	prefix        string
	excludePrefix []string
	excludeRegexp []*regexp.Regexp
	excludeSuffix []string
	allfiles      map[string]bool
}

func (wft *walkFileTree) setPrefix(prefix string) {
	wft.prefix = prefix
}

func (wft *walkFileTree) isExclude(fPath string) bool {
	if fPath == "" {
		return true
	}

	for _, prefix := range wft.excludePrefix {
		if strings.HasPrefix(fPath, prefix) {
			return true
		}
	}
	for _, suffix := range wft.excludeSuffix {
		if strings.HasSuffix(fPath, suffix) {
			return true
		}
	}
	return false
}

func (wft *walkFileTree) isExcludeName(name string) bool {
	for _, r := range wft.excludeRegexp {
		if r.MatchString(name) {
			return true
		}
	}

	return false
}

func (wft *walkFileTree) isEmpty(fpath string) bool {
	fh, _ := os.Open(fpath)
	defer fh.Close()
	infos, _ := fh.Readdir(-1)
	for _, fi := range infos {
		fn := fi.Name()
		fp := path.Join(fpath, fn)
		if wft.isExclude(wft.virPath(fp)) {
			continue
		}
		if wft.isExcludeName(fn) {
			continue
		}
		if fi.Mode()&os.ModeSymlink > 0 {
			continue
		}
		if fi.IsDir() && wft.isEmpty(fp) {
			continue
		}
		return false
	}
	return true
}

func (wft *walkFileTree) relName(fpath string) string {
	name, _ := filepath.Rel(wft.prefix, fpath)
	return name
}

func (wft *walkFileTree) virPath(fpath string) string {
	name := fpath[len(wft.prefix):]
	if name == "" {
		return ""
	}
	name = name[1:]
	return name
}

func (wft *walkFileTree) readDir(dirname string) ([]os.FileInfo, error) {
	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	list, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	sort.Sort(byName(list))
	return list, nil
}

func (wft *walkFileTree) walkLeaf(fpath string, fi os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if fpath == NewGGConfig().AppPath {
		return nil
	}

	if fi.IsDir() {
		return nil
	}

	if ssym && fi.Mode()&os.ModeSymlink > 0 {
		return nil
	}

	name := wft.virPath(fpath)

	if wft.allfiles[name] {
		return nil
	}

	if added, err := wft.wak.compress(name, fpath, fi); added {
		log.Printf("Compressed: %s\n", fpath)
		wft.allfiles[name] = true
		return err
	} else {
		log.Printf("Compressed fail: %s\n", name)
		return err
	}
}

func (wft *walkFileTree) iterDirectory(fpath string, fi os.FileInfo) error {
	doFSym := fsym && fi.Mode()&os.ModeSymlink > 0
	if doFSym {
		nfi, err := os.Stat(fpath)
		if os.IsNotExist(err) {
			return nil
		}
		fi = nfi
	}

	relPath := wft.virPath(fpath)

	if len(relPath) > 0 {
		if wft.isExcludeName(fi.Name()) {
			return nil
		}

		if wft.isExclude(relPath) {
			return nil
		}
	}

	err := wft.walkLeaf(fpath, fi, nil)
	if err != nil {
		if fi.IsDir() && err == filepath.SkipDir {
			return nil
		}
		return err
	}

	if !fi.IsDir() {
		return nil
	}

	list, err := wft.readDir(fpath)
	if err != nil {
		return wft.walkLeaf(fpath, fi, err)
	}

	for _, fileInfo := range list {
		err = wft.iterDirectory(path.Join(fpath, fileInfo.Name()), fileInfo)
		if err != nil {
			if !fileInfo.IsDir() || err != filepath.SkipDir {
				return err
			}
		}
	}
	return nil
}

func (wft *walkFileTree) walkRoot(root string) error {
	wft.prefix = root
	fi, err := os.Stat(root)
	if err != nil {
		return err
	}
	return wft.iterDirectory(root, fi)
}

type tarWalk struct {
	walkFileTree
	tw *tar.Writer
}

func (wft *tarWalk) compress(name, fpath string, fi os.FileInfo) (bool, error) {
	isSym := fi.Mode()&os.ModeSymlink > 0
	link := ""
	if isSym {
		link, _ = os.Readlink(fpath)
	}

	hdr, err := tar.FileInfoHeader(fi, link)
	if err != nil {
		return false, err
	}
	hdr.Name = name

	tw := wft.tw
	err = tw.WriteHeader(hdr)
	if err != nil {
		return false, err
	}

	if isSym == false {
		fr, err := os.Open(fpath)
		if err != nil {
			return false, err
		}
		defer fr.Close()
		_, err = io.Copy(tw, fr)
		if err != nil {
			return false, err
		}
		tw.Flush()
	}

	return true, nil
}

type zipWalk struct {
	walkFileTree
	zw *zip.Writer
}

func (wft *zipWalk) compress(name, fpath string, fi os.FileInfo) (bool, error) {
	isSym := fi.Mode()&os.ModeSymlink > 0

	hdr, err := zip.FileInfoHeader(fi)
	if err != nil {
		return false, err
	}
	hdr.Name = name

	zw := wft.zw
	w, err := zw.CreateHeader(hdr)
	if err != nil {
		return false, err
	}

	if isSym == false {
		fr, err := os.Open(fpath)
		if err != nil {
			return false, err
		}
		defer fr.Close()
		_, err = io.Copy(w, fr)
		if err != nil {
			return false, err
		}
	} else {
		var link string
		if link, err = os.Readlink(fpath); err != nil {
			return false, err
		}
		_, err = w.Write([]byte(link))
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

func packDirectory(excludePrefix []string, excludeSuffix []string,
	excludeRegexp []*regexp.Regexp, includePath ...string) (err error) {

	fmt.Printf("exclude relpath prefix: %s\n", strings.Join(excludePrefix, ":"))
	fmt.Printf("exclude relpath suffix: %s\n", strings.Join(excludeSuffix, ":"))
	if len(excludeRegexp) > 0 {
		fmt.Printf("exclude filename regex: `%v`\n", excludeRegexp)
	}
	log.Println("Create compressed file: ", NewGGConfig().AppPath)
	w, err := os.OpenFile(NewGGConfig().AppPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	var wft walker

	if NewGGConfig().PackFormat == "zip" {
		walk := new(zipWalk)
		zw := zip.NewWriter(w)
		defer func() {
			zw.Close()
		}()
		walk.allfiles = make(map[string]bool)
		walk.zw = zw
		walk.wak = walk
		walk.excludePrefix = excludePrefix
		walk.excludeSuffix = excludeSuffix
		walk.excludeRegexp = excludeRegexp
		wft = walk
	} else {
		walk := new(tarWalk)
		cw := gzip.NewWriter(w)
		tw := tar.NewWriter(cw)

		defer func() {
			tw.Flush()
			cw.Flush()
			tw.Close()
			cw.Close()
		}()
		walk.allfiles = make(map[string]bool)
		walk.tw = tw
		walk.wak = walk
		walk.excludePrefix = excludePrefix
		walk.excludeSuffix = excludeSuffix
		walk.excludeRegexp = excludeRegexp
		wft = walk
	}

	for _, p := range includePath {
		err = wft.walkRoot(p)
		if err != nil {
			log.Println("Tar error:", err)
			return
		}
	}

	return
}

func Pack() (err error) {
	return packDirectory(NewGGConfig().PackExcludePrefix, NewGGConfig().PackExcludeSuffix, NewGGConfig().PackExcludeRegexp, NewGGConfig().PackPaths...)
}

func unPackFile(fileName string) (err error) {
	log.Println("Tar package ", fileName)
	// file read
	fr, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer fr.Close()
	// gzip read
	gr, err := gzip.NewReader(fr)
	if err != nil {
		return err
	}
	defer gr.Close()
	// tar read
	tr := tar.NewReader(gr)
	// 读取文件
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		// Skip the path of .
		if h.Name == "" {
			continue
		}
		// 显示文件
		log.Println("Uncompressed:", h.Name)
		// 打开文件
		tar_path := (NewGGConfig().RunDirectory + h.Name)[:strings.LastIndex(NewGGConfig().RunDirectory+h.Name, "/")]
		if err := os.MkdirAll(tar_path, os.ModePerm); err != nil {
			log.Println("Mkdir error: ", tar_path)
		}
		fw, err := os.OpenFile(NewGGConfig().RunDirectory+h.Name, os.O_CREATE|os.O_WRONLY, 0777 /*os.FileMode(h.Mode)*/)
		if err != nil {
			return err
		}
		defer fw.Close()
		// 写文件
		_, err = io.Copy(fw, tr)
		if err != nil {
			return err
		}
	}
	return nil
}

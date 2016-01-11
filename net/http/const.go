package http

const (
	// CR: 13 LF: 10
	CRLF            = "\r\n"
	LOG_TIME_FORMAT = "2/Jan/2006:15:04:05 -0700"

	/*
		log_format compression '$remote_addr - $remote_user [$time_local] '
		'"$request" $status $bytes_sent '
		'"$http_referer" "$http_user_agent" "$gzip_ratio"';
	*/
	LOG_CONTEXT = `%s - %s [%s] "%s %s %s" %d %d "%s" "%s" "%d"`

	CYEAM_LOG = `
   ____ _____   _   _   _
  / _\ V / __| / \ | \_/ |  
 ( (_ \ /| _| | o || \_/ |
  \__||_||___||_n_||_| |_| 
 `
)

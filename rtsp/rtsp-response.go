package rtsp

import (
	"fmt"
	"strconv"
)

const (
	RTSP_VERSION = "RTSP/1.0"
)

type RtspResponse struct {
	Version    string
	StatusCode int
	Status     string
	Header     map[string]interface{}
	Body       string
}

func NewRtspResponse(statusCode int, status, cSeq, sid, body string) *RtspResponse {
	res := &RtspResponse{
		Version:    RTSP_VERSION,
		StatusCode: statusCode,
		Status:     status,
		Header:     map[string]interface{}{"CSeq": cSeq, "Session": sid},
		Body:       body,
	}
	len := len(body)
	if len > 0 {
		res.Header["Content-Length"] = strconv.Itoa(len)
	} else {
		delete(res.Header, "Content-Length")
	}
	return res
}

func (r *RtspResponse) String() string {
	str := fmt.Sprintf("%s %d %s\r\n", r.Version, r.StatusCode, r.Status)
	for key, value := range r.Header {
		str += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	str += "\r\n"
	str += r.Body
	return str
}

func (r *RtspResponse) SetBody(body string) {
	len := len(body)
	r.Body = body
	if len > 0 {
		r.Header["Content-Length"] = strconv.Itoa(len)
	} else {
		delete(r.Header, "Content-Length")
	}
}

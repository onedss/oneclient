package rtsp

import (
	"bufio"
	"fmt"
	"github.com/reactivex/rxgo/observable"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
)

type RtspClient struct {
	Stoped      bool
	Path        string
	Conn        *net.Conn
	AuthHeaders bool
	Session     *string
	Seq         int
	connRW      *bufio.ReadWriter
	InBytes     uint64
}

func NewRtspClient(path string) *RtspClient {
	session := &RtspClient{
		Stoped: false,
		Path:   path,
	}
	return session
}

func (client *RtspClient) Start() observable.Observable {
	return observable.Start(func() interface{} {
		l, err := url.Parse(client.Path)
		if err != nil {
			return err
		}
		conn, err := net.Dial("tcp", l.Hostname()+":"+l.Port())
		if err != nil {
			// handle error
			return err
		}
		client.Conn = &conn
		client.connRW = bufio.NewReadWriter(bufio.NewReaderSize(conn, 10240), bufio.NewWriterSize(conn, 10240))

		headers := make(map[string]string)
		headers["Require"] = "implicit-play"
		// An OPTIONS request returns the request types the server will accept.
		resp, err := client.Request("OPTIONS", headers)
		if err != nil {
			return err
		}
		fmt.Println("StatusCode:", resp.StatusCode)

		// A DESCRIBE request includes an RTSP URL (rtsp://...), and the type of reply data that can be handled. This reply includes the presentation description,
		// typically in Session Description Protocol (SDP) format. Among other things, the presentation description lists the media streams controlled with the aggregate URL.
		// In the typical case, there is one media stream each for audio and video.
		headers = make(map[string]string)
		headers["Accept"] = "application/sdp"
		resp, err = client.Request("DESCRIBE", headers)
		if err != nil {
			return err
		}

		//fmt.Println("StatusCode:",resp.StatusCode)
		headers = make(map[string]string)
		headers["Transport"] = "RTP/AVP;unicast;client_port=8000-8001"
		resp, err = client.Request("SETUP", headers)
		if err != nil {
			return err
		}

		//fmt.Println("StatusCode:",resp.StatusCode)

		////fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
		////status, err := bufio.NewReader(conn).ReadString('\n')
		////url.Host
		////text, err := reader.ReadString('\n')
		////if err != nil {
		////	return err
		////}
		//return text

		return 0
	})
	//return observable.Just(1)
}

func (client *RtspClient) Request(method string, headers map[string]string) (resp *RtspResponse, err error) {
	headers["User-Agent"] = "OneDarwinGo"
	if client.AuthHeaders {
		//headers["Authorization"] = this.digest(method, _url);
	}
	if client.Session != nil {
		headers["Session"] = *client.Session
	}
	client.Seq++
	cseq := client.Seq
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("%s %s RTSP/1.0\r\n", method, client.Path))
	builder.WriteString(fmt.Sprintf("CSeq: %d\r\n", cseq))
	for k, v := range headers {
		builder.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	builder.WriteString(fmt.Sprintf("\r\n"))
	s := builder.String()
	fmt.Println("C->S	>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
	fmt.Println(s)
	_, err = client.connRW.WriteString(s)
	if err != nil {
		return
	}
	client.connRW.Flush()
	lineCount := 0
	statusCode := 200
	status := ""
	sid := ""
	contentLen := 0
	respHeader := make(map[string]string)
	var line []byte
	builder.Reset()
	for !client.Stoped {
		if line, _, err = client.connRW.ReadLine(); err != nil {
			return
		} else {
			if len(line) == 0 {
				fmt.Println("S->C	<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<")
				fmt.Println(builder.String())
				resp = NewRtspResponse(statusCode, status, strconv.Itoa(cseq), sid, "")
				return
			}
			s := string(line)
			builder.Write(line)
			builder.WriteString("\r\n")

			if lineCount == 0 {
				splits := strings.Split(s, " ")
				if len(splits) < 3 {
					err = fmt.Errorf("StatusCode Line error:%s", s)
					return
				}
				statusCode, err = strconv.Atoi(splits[1])
				if err != nil {
					return
				}
				if statusCode != 200 {
					err = fmt.Errorf("Response StatusCode is :%d", statusCode)
					return
				}
				status = splits[2]
			}
			lineCount++
			splits := strings.Split(s, ":")
			if len(splits) == 2 {
				respHeader[splits[0]] = strings.TrimSpace(splits[1])
			}
			if strings.Index(s, "Session:") == 0 {
				splits := strings.Split(s, ":")
				sid = strings.TrimSpace(splits[1])
			}
			//if strings.Index(s, "CSeq:") == 0 {
			//	splits := strings.Split(s, ":")
			//	cseq, err = strconv.Atoi(strings.TrimSpace(splits[1]))
			//	if err != nil {
			//		err = fmt.Errorf("Atoi CSeq err. line:%s", s)
			//		return
			//	}
			//}
			if strings.Index(s, "Content-Length:") == 0 {
				splits := strings.Split(s, ":")
				contentLen, err = strconv.Atoi(strings.TrimSpace(splits[1]))
				if err != nil {
					return
				}
				content := make([]byte, contentLen)
				_, err = io.ReadFull(client.connRW, content)
				if err != nil {
					err = fmt.Errorf("Read content err.ContentLength:%d", contentLen)
					return
				}
				body := string(content)
				builder.Write(content)
				resp = &RtspResponse{
					Body:       body,
					Status:     status,
					StatusCode: statusCode,
				}

				fmt.Println("S->C	<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<")
				fmt.Println(builder.String())
				return
			}
		}
	}
	return
}

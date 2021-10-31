package main

import (
	"fmt"
	"github.com/onedss/EasyGoLib/utils"
	"github.com/onedss/oneclient/rtsp"
)

func main() {
	client := rtsp.NewRtspClient("rtsp://47.96.132.39:554/45326250000000101010101202110300001.sdp")
	observable := client.Start()
	if elem, err := <-observable; err {
		fmt.Println("Received", elem)
	}
	utils.PauseExit()
}

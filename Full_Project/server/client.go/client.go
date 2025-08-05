package main

import (
	"encoding/base64"
	"fmt"
	"image"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"gocv.io/x/gocv"
)

const (
	droneID     = "drone_1" // unique ID
	videoFile   = "A.mp4"   // your test video
	jpegQuality = 70        // reduce for lower latency
	fps         = 15        // streaming FPS
)

func main() {
	fmt.Println("Starting stream for", droneID)

	webcam, err := gocv.VideoCaptureFile(videoFile)
	if err != nil {
		log.Fatal("Error opening video file:", err)
	}
	defer webcam.Close()

	img := gocv.NewMat()
	defer img.Close()

	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/stream", nil)
	if err != nil {
		log.Fatal("WebSocket dial error:", err)
	}
	defer conn.Close()

	ticker := time.NewTicker(time.Second / fps)
	defer ticker.Stop()

	for range ticker.C {
		if ok := webcam.Read(&img); !ok || img.Empty() {
			webcam.Set(gocv.VideoCapturePosFrames, 0) // loop video
			continue
		}

		// Resize for better performance
		gocv.Resize(img, &img, image.Pt(0, 0), 0.3, 0.3, gocv.InterpolationArea)

		buf, err := gocv.IMEncodeWithParams(".jpg", img, []int{gocv.IMWriteJpegQuality, jpegQuality})
		if err != nil {
			continue
		}
		base64Str := base64.StdEncoding.EncodeToString(buf.GetBytes())
		buf.Close()

		msg := droneID + "|" + base64Str
		conn.WriteMessage(websocket.TextMessage, []byte(msg))
	}
}

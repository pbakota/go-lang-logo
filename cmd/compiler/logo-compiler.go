package main

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"

	"rs.lab/go-logo/logo"
)

const TEMPLATE = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Logo</title>
	<style>
	canvas {
		padding-left: 0;
		padding-right: 0;
		margin-left: auto;
		margin-right: auto;
		display: block;
		width: 640px;
	}	
	</style>
</head>
<body>
    <canvas width="640" height="480" id="canvas"></canvas>
    <script>
        const canvas = document.getElementById('canvas');
        const ctx = canvas.getContext("2d");
		// To get clear line without AA
		ctx.translate(0.5, 0.5);

        var paper = 'black';
        var ink = 'white';
        var head = {x: 320, y: 240, angle: 0};
        var pendown = false;

        
        const clear = () => {
            ctx.fillStyle = paper;
            ctx.fillRect(0,0,canvas.width, canvas.height);
        }

        const home = () => {
            var head = {x: 320, y: 240, angle: 0};
            var pendown = false;
            clear();
        }
        
        const drawLine = (x1, y1, x2, y2) => {
            ctx.strokeStyle = ink;
            ctx.beginPath();
            ctx.moveTo(x1,y1);
            ctx.lineTo(x2,y2);
            ctx.stroke();
        }
        
        const degToRad = (deg) => deg * (Math.PI / 180);

        const calcOffset = (step) => {
            const dx = step * Math.cos(degToRad(head.angle))
            const dy = step * Math.sin(degToRad(head.angle))
            return {dx: dx, dy: dy};
        }

        const forward = (step) => {
            const {dx, dy} = calcOffset(step);
            if (pendown) {
		        drawLine(head.x, head.y, head.x+dx, head.y+dy)
            }
            head.x += dx
            head.y += dy
        }

        const back = (step) => {
            const {dx, dy} = calcOffset(step);
            if (pendown) {
		        drawLine(head.x, head.y, head.x-dx, head.y-dy)
            }
            head.x -= dx
            head.y -= dy
        }

        const left = (value) => {
            head.angle = (head.angle + value) % 360;
        }

        const right = (value) => {
            head.angle = (head.angle - value) % 360;
        }


        // {{compiled-code}}
    </script>
</body>
</html>
`

func main() {
	source, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	buffer := bytes.Buffer{}
	writer := bufio.NewWriter(&buffer)

	c := logo.NewCompiler(writer)
	// c.Trace = true
	err = c.Run(string(source))
	if err != nil {
		panic(err)
	}
	writer.Flush()

	compiled := buffer.String()
	os.Stdout.WriteString(strings.Replace(TEMPLATE, "// {{compiled-code}}", compiled, -1))
}

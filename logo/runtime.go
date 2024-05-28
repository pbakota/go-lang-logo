package logo

import (
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
)

type Color int
type Command func(r *Runtime)

type DrawingStub interface {
	Clear(r *Runtime)
	DrawLine(r *Runtime, x1, y1, x2, y2 int32)
}

type NullDraw struct {
	DrawingStub
}

func NewNullInterpreter() *NullDraw {
	return &NullDraw{}
}

func (i *NullDraw) DrawLine(r *Runtime, x1, y1, x2, y2 int32) {}
func (i *NullDraw) Clear(r *Runtime)                          {}

type Position struct {
	X, Y int32
}

type ProgramStep struct {
	Token   Token
	Line    uint32
	String  string
	Number  int
	Literal rune
}

type Runtime struct {
	Program []ProgramStep
	Stub    DrawingStub
	PC      int
	Stack   [256]int // this allow 128 nested loops
	SP      int
	Trace   bool

	Head    Position
	Angle   int
	PenDown bool
	Paper   Color
	Ink     Color
}

const (
	Black   = iota
	White   = iota
	Red     = iota
	Green   = iota
	Blue    = iota
	Yellow  = iota
	Gray    = iota
	Magenta = iota
)

var COLORS = map[string]Color{
	"BLACK":   Black,
	"WHITE":   White,
	"RED":     Red,
	"GREEN":   Green,
	"BLUE":    Blue,
	"YELLOW":  Yellow,
	"GRAY":    Gray,
	"MAGENTA": Magenta,
}

var KEYWORDS = map[string]Command{
	"HOME":    homeCmd,
	"PAPER":   paperCmd,
	"INK":     inkCmd,
	"PEN":     penCmd,
	"REPEAT":  repeatCmd,
	"LOOP":    loopCmd,
	"FORWARD": forwardCmd,
	"BACK":    backCmd,
	"LEFT":    leftCmd,
	"RIGHT":   rightCmd,
}

func homeCmd(r *Runtime) {
	r.trace("HOME")
	r.Head = Position{X: 320, Y: 240}
	r.Stub.Clear(r)
}

func paperCmd(r *Runtime) {
	r.trace("PAPER")
	r.Paper = r.getColor()
}

func inkCmd(r *Runtime) {
	r.trace("INK")
	r.Ink = r.getColor()
}

func penCmd(r *Runtime) {
	r.trace("PEN")
	token := r.getParam(TkIdent)
	value := strings.ToUpper(token.String)
	if value == "UP" || value == "DOWN" {
		r.PenDown = value == "DOWN"
		return
	}

	r.syntaxError(fmt.Sprintf("invalid parameter in line %d", token.Line))
}

func forwardCmd(r *Runtime) {
	r.trace("FORWARD")
	step := r.getParam(TkNumber).Number
	dx := int32(math.Round(float64(step) * math.Cos(r.DegToRad(r.Angle))))
	dy := int32(math.Round(float64(step) * math.Sin(r.DegToRad(r.Angle))))

	if r.PenDown {
		r.Stub.DrawLine(r, r.Head.X, r.Head.Y, r.Head.X+dx, r.Head.Y+dy)
	}

	r.Head.X += dx
	r.Head.Y += dy
}

func backCmd(r *Runtime) {
	r.trace("BACK")
	step := r.getParam(TkNumber).Number
	dx := int32(float64(step) * math.Cos(r.DegToRad(r.Angle)))
	dy := int32(float64(step) * math.Sin(r.DegToRad(r.Angle)))

	if r.PenDown {
		r.Stub.DrawLine(r, r.Head.X, r.Head.Y, r.Head.X-dx, r.Head.Y-dy)
	}

	r.Head.X -= dx
	r.Head.Y -= dy
}

func leftCmd(r *Runtime) {
	r.trace("LEFT")
	r.Angle = (r.Angle + r.getParam(TkNumber).Number) % 360
}

func rightCmd(r *Runtime) {
	r.trace("RIGHT")
	r.Angle = (r.Angle - r.getParam(TkNumber).Number) % 360
}

func repeatCmd(r *Runtime) {
	r.trace("REPEAT")
	token := r.getParam(TkNumber)
	if token.Number <= 0 || token.Number >= 65536 {
		r.syntaxError(fmt.Sprintf("the count is too small or too large number in line %d", token.Line))
	}

	r.push(r.PC)         // save pc
	r.push(token.Number) // save counter
}

func loopCmd(r *Runtime) {
	r.trace("LOOP")
	value := r.pop()
	value -= 1
	if value > 0 {
		pc := r.pop()
		r.push(pc)
		r.push(value)
		r.PC = pc // restore pc
	} else {
		r.pop() // remove pc
	}
}

func (r *Runtime) trace(msg string) {
	if r.Trace {
		log.Printf("TRACE: %s\n", msg)
	}
}

func (r *Runtime) runtimeError(err error) {
	panic(err)
}

func (r *Runtime) syntaxError(err string) {
	r.runtimeError(fmt.Errorf("syntax error: %s", err))
}

func (r *Runtime) getColor() Color {
	token := r.getParam(TkIdent)
	color := strings.ToUpper(token.String)
	if value, ok := COLORS[color]; ok {
		return value
	}
	r.syntaxError(fmt.Sprintf("unrecognized color in line %d", token.Line))
	return Black // Dummy color
}

func (r *Runtime) isEOP() bool {
	return r.PC == int(len(r.Program))
}

func (r *Runtime) expectToken() {
	if r.isEOP() {
		r.runtimeError(errors.New("unexpected end of program"))
	}
}

func (r *Runtime) next() ProgramStep {
	r.expectToken()
	token := r.Program[r.PC]
	r.PC += 1
	return token
}

func (r *Runtime) getParam(expected Token) ProgramStep {
	param := r.next()
	if param.Token != expected {
		r.syntaxError(fmt.Sprintf("invalid parameter type, expected %d got %d in line %d", expected, param.Token, param.Line))
	}

	return param
}

func (r *Runtime) push(val int) {
	if r.SP == len(r.Stack) {
		r.runtimeError(errors.New("stack overflow"))
	}
	r.Stack[r.SP] = val
	r.SP += 1
}

func (r *Runtime) pop() int {
	if r.SP == 0 {
		r.runtimeError(errors.New("stack empty"))
	}
	r.SP -= 1
	return r.Stack[r.SP]
}

func NewRuntime() *Runtime {
	return &Runtime{
		Stub:    NewNullInterpreter(),
		PC:      0,
		SP:      0,
		Head:    Position{X: 320, Y: 240},
		Paper:   Black,
		Ink:     White,
		Program: []ProgramStep{},
		Trace:   false,
	}
}

func (r *Runtime) DegToRad(deg int) float64 {
	return float64(deg) * (math.Pi / 180)
}

func (r *Runtime) Run(program string) error {
	l := NewLexer(program)
	// l.Debug = true
	r.Program = []ProgramStep{}

	// Parsing the source, building the program steps
	for {
		token, err := l.NextToken()
		if err != nil {
			return err
		}

		if token == TkEOF {
			break
		}

		switch token {
		case TkIdent:
			r.Program = append(r.Program, ProgramStep{Token: token, String: l.String, Line: l.Line})
		case TkNumber:
			r.Program = append(r.Program, ProgramStep{Token: token, Number: l.Number, Line: l.Line})
		case TkLiteral: // skipped
		case TkEOL: // skipped
		case TkComment: // skipped
		default:
			return fmt.Errorf("invalid token %d in line %d", token, l.Line)
		}
	}

	// Running the program
	r.PC = 0 // reset
	for {
		p := r.next()

		if p.Token != TkIdent {
			r.syntaxError(fmt.Sprintf("unexpected token %d in line %d", p.Token, p.Line))
		}

		cmd := strings.ToUpper(p.String)

		if fn, ok := KEYWORDS[cmd]; ok {
			fn(r)
		} else {
			r.syntaxError(fmt.Sprintf("unknown keyword in line %d", p.Line))
		}

		if r.isEOP() {
			break
		}
	}

	return nil
}

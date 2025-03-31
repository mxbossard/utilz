package printz

import (
	"fmt"
	"io"
	"log"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mxbossard/utilz/anzi"
	"github.com/mxbossard/utilz/errorz"
	"github.com/mxbossard/utilz/formatz"
)

// Printer responsible for printing messages in outputs (example: print with colors, without colors, ...)

type Flusher interface {
	Flush() error
}

type Closer interface {
	Close(messages string) error
}

type Printer interface {
	Flusher
	Flushed() bool
	RecoverableOut(...interface{}) error
	Out(...interface{})
	Outf(string, ...interface{})
	ColoredOutf(anzi.Color, string, ...interface{})
	RecoverableErr(...interface{}) error
	Err(...interface{})
	Errf(string, ...interface{})
	ColoredErrf(anzi.Color, string, ...interface{})
	LastPrint() time.Time
	Outputs() Outputs
}

type ClosingPrinter interface {
	Printer
	Closer
	IsClosed() bool
}

type basicPrinter struct {
	*sync.Mutex
	outputs   Outputs
	lastPrint time.Time
}

func (p *basicPrinter) Outputs() Outputs {
	return p.outputs
}

func (o basicPrinter) Flush() error {
	o.Lock()
	defer o.Unlock()
	err := o.outputs.Flush()
	return err
}

func (o basicPrinter) Flushed() bool {
	return o.outputs.Flushed()
}

func (p *basicPrinter) RecoverableOut(objects ...interface{}) (err error) {
	p.Lock()
	defer p.Unlock()
	p.lastPrint = time.Now()
	//_, err = fmt.Fprint(p.outputs.Out(), objects...)
	err = printTo(p.outputs.Out(), objects...)
	return
}

func (p *basicPrinter) Out(objects ...interface{}) {
	err := p.RecoverableOut(objects...)
	if err != nil {
		log.Fatal(err)
	}
}

func (p *basicPrinter) Outf(s string, params ...interface{}) {
	s = fmt.Sprintf(s, params...)
	p.Out(s)
}

func (p *basicPrinter) ColoredOutf(color anzi.Color, s string, params ...interface{}) {
	s = formatz.Sprintf(color, s, ansiFormatParams(color, params...)...)
	p.Out(s)
}

func (p *basicPrinter) RecoverableErr(objects ...interface{}) (err error) {
	p.Lock()
	defer p.Unlock()
	p.lastPrint = time.Now()
	//_, err = fmt.Fprint(p.outputs.Err(), objects...)
	err = printTo(p.outputs.Err(), objects...)
	return
}

func (p *basicPrinter) Err(objects ...interface{}) {
	err := p.RecoverableErr(objects...)
	if err != nil {
		log.Fatal(err)
	}
}

func (p *basicPrinter) Errf(s string, params ...interface{}) {
	s = fmt.Sprintf(s, params...)
	p.Err(s)
}

func (p *basicPrinter) ColoredErrf(color anzi.Color, s string, params ...interface{}) {
	s = formatz.Sprintf(color, s, ansiFormatParams(color, params...)...)
	p.Err(s)
}

func (p basicPrinter) LastPrint() time.Time {
	if p.outputs.LastPrint().After(p.lastPrint) {
		return p.outputs.LastPrint()
	}
	return p.lastPrint
}

func stringify(obj interface{}) (str string, err error) {
	switch o := obj.(type) {
	case string:
		str = o
	case int:
		str = strconv.Itoa(o)
	case float64:
		str = strconv.FormatFloat(o, 'E', 3, 32)

	case ansiFormatted:
		if o.Content == "" {
			return "", nil
		}
		content, err := stringify(o.Content)
		if err != nil {
			return "", err
		}

		if o.LeftPad > 0 {
			spaceCount := o.LeftPad - len(content)
			if spaceCount > 0 {
				content = strings.Repeat(" ", spaceCount) + content
			}
		} else if o.RightPad > 0 {
			spaceCount := o.RightPad - len(content)
			if spaceCount > 0 {
				content += strings.Repeat(" ", spaceCount)
			}
		}

		if o.Format != "" {
			content = fmt.Sprintf("%s%s%s", o.Format, content, anzi.Reset)
		}

		str = content
	case error:
		str = fmt.Sprintf("Error: %s !\n", obj)
	default:
		err = fmt.Errorf("unable to Print object of type: %T", obj)
		return
	}
	return
}

func printTo(w io.Writer, objects ...interface{}) (err error) {
	objects = expandObjects(objects)
	var toPrint []string

	for _, obj := range objects {
		switch o := obj.(type) {
		case errorz.Aggregated:
			for _, err := range o.Errors() {
				//fmt.Fprintf(w, "Error: %s !\n", err)
				printTo(w, err)
			}
		default:
			str, err := stringify(o)
			if err != nil {
				return err
			}
			if len(str) > 0 {
				//fmt.Printf("adding str: %d [%s](%T)\n", len(str), str, str)
				toPrint = append(toPrint, str)
			}
		}
	}
	//fmt.Printf("toPrint: %d %s\n", len(toPrint), toPrint)
	_, err = fmt.Fprintf(w, "%s", strings.Join(toPrint, ""))
	return
}

func expandObjects(objects ...interface{}) (allObjects []interface{}) {
	for _, obj := range objects {
		// Recursive call if obj is an array or a slice
		t := reflect.TypeOf(obj)
		if t.Kind() == reflect.Array || t.Kind() == reflect.Slice {
			arrayValue := reflect.ValueOf(obj)
			for i := 0; i < arrayValue.Len(); i++ {
				value := arrayValue.Index(i).Interface()
				expanded := expandObjects(value)
				allObjects = append(allObjects, expanded...)
			}
			continue
		} else {
			allObjects = append(allObjects, obj)
		}
	}
	return
}

type closingPrinter struct {
	Printer

	closed  bool
	message string
}

func (p *closingPrinter) Close(message string) (err error) {
	if p.closed {
		return fmt.Errorf("printer already closed with message: [%s]", p.message)
	}
	p.closed = true
	p.message = message
	return
}

func (p *closingPrinter) IsClosed() bool {
	return p.closed
}

func (p *closingPrinter) RecoverableOut(objects ...interface{}) (err error) {
	if p.closed {
		return fmt.Errorf("printer already closed with message: [%s]", p.message)
	}
	return p.Printer.RecoverableOut(objects...)
}

func (p *closingPrinter) Out(objects ...interface{}) {
	if p.closed {
		panic(fmt.Errorf("printer already closed with message: [%s]", p.message))
	}
	p.Printer.Out(objects...)
}

func (p *closingPrinter) Outf(s string, params ...interface{}) {
	if p.closed {
		panic(fmt.Errorf("printer already closed with message: [%s]", p.message))
	}
	p.Printer.Outf(s, params...)
}

func (p *closingPrinter) ColoredOutf(color anzi.Color, s string, params ...interface{}) {
	if p.closed {
		panic(fmt.Errorf("printer already closed with message: [%s]", p.message))
	}
	p.Printer.ColoredOutf(color, s, params...)
}

func (p *closingPrinter) RecoverableErr(objects ...interface{}) (err error) {
	if p.closed {
		return fmt.Errorf("printer already closed with message: [%s]", p.message)
	}
	return p.Printer.RecoverableErr(objects...)
}

func (p *closingPrinter) Err(objects ...interface{}) {
	if p.closed {
		panic(fmt.Errorf("printer already closed with message: [%s]", p.message))
	}
	p.Printer.Err(objects...)
}

func (p *closingPrinter) Errf(s string, params ...interface{}) {
	if p.closed {
		panic(fmt.Errorf("printer already closed with message: [%s]", p.message))
	}
	p.Printer.Errf(s, params...)
}

func (p *closingPrinter) ColoredErrf(color anzi.Color, s string, params ...interface{}) {
	if p.closed {
		panic(fmt.Errorf("printer already closed with message: [%s]", p.message))
	}
	p.Printer.ColoredErrf(color, s, params...)
}

// Build a default printer with buffered outputs.
func New(outputs Outputs) Printer {
	buffered := NewBufferedOutputs(outputs)
	var m sync.Mutex
	var t time.Time
	printer := basicPrinter{&m, buffered, t}

	return &printer
}

func NewUnbuffured(outputs Outputs) Printer {
	var m sync.Mutex
	var t time.Time
	printer := basicPrinter{&m, outputs, t}
	return &printer
}

// Build a default printer with non buffered standards outputs.
func NewStandard() Printer {
	outputs := NewStandardOutputs()
	return New(outputs)
}

// Build a default printer with discarding outputs.
func NewDiscarding() Printer {
	outputs := NewOutputs(io.Discard, io.Discard)
	return New(outputs)
}

func Buffered(p Printer) Printer {
	buffered := NewBufferedOutputs(p.Outputs())
	var m sync.Mutex
	var t time.Time
	printer := basicPrinter{&m, buffered, t}
	return &printer
}

func Closing(p Printer) ClosingPrinter {
	closing := closingPrinter{Printer: p}
	return &closing
}

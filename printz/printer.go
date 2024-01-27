package printz

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"mby.fr/utils/ansi"
	"mby.fr/utils/errorz"
)

// Printer responsible for printing messages in outputs (example: print with colors, without colors, ...)

type Flusher interface {
	Flush() error
}

type Printer interface {
	Outputs() Outputs
	Flush() error
	RecoverableOut(...interface{}) error
	Out(...interface{})
	Outf(string, ...interface{})
	ColoredOutf(ansi.Color, string, ...interface{})
	RecoverableErr(...interface{}) error
	Err(...interface{})
	Errf(string, ...interface{})
	ColoredErrf(ansi.Color, string, ...interface{})
	//Print(...interface{}) error
	LastPrint() time.Time
}

type BasicPrinter struct {
	sync.Mutex
	outputs   Outputs
	lastPrint time.Time
}

func (p *BasicPrinter) Outputs() Outputs {
	return p.outputs
}

func (o BasicPrinter) Flush() error {
	o.Lock()
	defer o.Unlock()
	return o.outputs.Flush()
}

func (p *BasicPrinter) RecoverableOut(objects ...interface{}) (err error) {
	p.Lock()
	defer p.Unlock()
	p.lastPrint = time.Now()
	//_, err = fmt.Fprint(p.outputs.Out(), objects...)
	err = printTo(p.outputs.Out(), objects...)
	return
}

func (p *BasicPrinter) Out(objects ...interface{}) {
	err := p.RecoverableOut(objects...)
	if err != nil {
		log.Fatal(err)
	}
}

func (p *BasicPrinter) Outf(s string, params ...interface{}) {
	s = fmt.Sprintf(s, params...)
	p.Out(s)
}

func (p *BasicPrinter) ColoredOutf(color ansi.Color, s string, params ...interface{}) {
	s = ansi.Sprintf(color, s, params...)
	p.Out(s)
}

func (p *BasicPrinter) RecoverableErr(objects ...interface{}) (err error) {
	p.Lock()
	defer p.Unlock()
	p.lastPrint = time.Now()
	//_, err = fmt.Fprint(p.outputs.Err(), objects...)
	err = printTo(p.outputs.Err(), objects...)
	return
}

func (p *BasicPrinter) Err(objects ...interface{}) {
	err := p.RecoverableErr(objects...)
	if err != nil {
		log.Fatal(err)
	}
}

func (p *BasicPrinter) Errf(s string, params ...interface{}) {
	s = fmt.Sprintf(s, params...)
	p.Err(s)
}

func (p *BasicPrinter) ColoredErrf(color ansi.Color, s string, params ...interface{}) {
	s = ansi.Sprintf(color, s, params...)
	p.Err(s)
}

func (p BasicPrinter) LastPrint() time.Time {
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

	/*
	case ansiFormatted:
		if o.content == "" {
			return "", nil
		}
		content, err := stringify(o.content)
		if err != nil {
			return "", err
		}
		if o.format != "" {
			str = fmt.Sprintf("%s%s%s", o.format, content, ansi.Reset)
		} else {
			str = content
		}
		if o.tab {
			str += "\t"
		} else if o.leftPad > 0 {
			spaceCount := o.leftPad - len(content)
			if spaceCount > 0 {
				str = strings.Repeat(" ", spaceCount) + str
			}
		} else if o.rightPad > 0 {
			spaceCount := o.rightPad - len(content)
			if spaceCount > 0 {
				str += strings.Repeat(" ", spaceCount)
			}
		}
	*/
	case error:
		str = fmt.Sprintf("Error: %s !\n", obj)
	default:
		err = fmt.Errorf("Unable to Print object of type: %T", obj)
		return
	}
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

func flush(printer *BasicPrinter) {
	printer.Lock()
	defer printer.Unlock()
	err := printer.outputs.Out().(*bufio.Writer).Flush()
	if err != nil {
		log.Fatal(err)
	}
	err = printer.outputs.Err().(*bufio.Writer).Flush()
	if err != nil {
		log.Fatal(err)
	}
}

//var flushablePrinters []*BasicPrinter

func New(outputs Outputs) Printer {
	buffered := NewBufferedOutputs(outputs)
	var m sync.Mutex
	var t time.Time
	printer := BasicPrinter{m, buffered, t}

	//flushablePrinters = append(flushablePrinters, &printer)

	// Flush printer every 5 seconds
	//go func() {
	//	for {
	//		time.Sleep(5 * time.Second)
	//		flush(&printer)
	//	}
	//}()

	return &printer
}

func NewStandard() Printer {
	outputs := NewStandardOutputs()
	return New(outputs)
}

// ANSI formatting for content
/*
type ansiFormatted struct {
	format            string
	content           interface{}
	tab               bool
	leftPad, rightPad int
}
*/
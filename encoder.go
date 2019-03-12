package gelfconv

import "io"

type Encoder struct {
	writer   io.Writer
	msgCount int
}

func NewEncoder(writer io.Writer) *Encoder {
	enc := Encoder{
		writer:   writer,
		msgCount: 0,
	}
	return &enc
}

func (x *Encoder) Write(msg Message) error {
	raw, err := msg.Gelf()
	if err != nil {
		return err
	}
	raw = append(raw, 0)

	for p := 0; p < len(raw); {
		if n, err := x.writer.Write(raw[p:len(raw)]); err != nil {
			return err
		} else {
			p += n
		}
	}

	x.msgCount++
	return nil
}

package target

import (
	"encoding/gob"
	"fmt"
	"strings"

	"github.com/portal-co/boogie/hashmap"
)

func EndsLabel(x string) bool {
	return !strings.HasPrefix(x, "//") || !strings.HasPrefix(x, "[") || x == ""
}

type Label interface {
	isLabel() struct{}
	fmt.Stringer
}
type IpfsLabel string

func (i IpfsLabel) String() string {
	return string(i) + "//"
}

func (i IpfsLabel) isLabel() struct{} {
	return struct{}{}
}

func (i *IpfsLabel) Parse(x *string) {
	for EndsLabel(*x) {
		r := []rune(*i)
		s := []rune(*x)
		r = append(r, s[0])
		s = s[1:]
		*i = IpfsLabel(r)
		*x = string(s)
	}
}

type BakedLabel string

func (b BakedLabel) isLabel() struct{} {
	return struct{}{}
}

func (b BakedLabel) String() string {
	return string(b) + "//"
}

func (b *BakedLabel) Parse(x *string) {
	for EndsLabel(*x) {
		r := []rune(*b)
		s := []rune(*x)
		r = append(r, s[0])
		s = s[1:]
		*b = BakedLabel(r)
		*x = string(s)
	}
}

type DelveLabel struct {
	Internal ConfiguredLabel
	Path     string
}

func (d DelveLabel) isLabel() struct{} {
	return struct{}{}
}

func (d *DelveLabel) Parse(x *string) {
	*x = strings.TrimPrefix(*x, "@")
	d.Internal.Parse(x)
	*x = strings.TrimPrefix(*x, "//")
	for EndsLabel(*x) {
		r := []rune(d.Path)
		s := []rune(*x)
		r = append(r, s[0])
		s = s[1:]
		d.Path = string(r)
		*x = string(s)
	}
}

func (d DelveLabel) String() string {
	return fmt.Sprintf("@%s%s//", d.Internal, d.Path)
}

func ParseLabel(x *string) Label {
	switch (*x)[0] {
	case 'Q':
		var l IpfsLabel
		l.Parse(x)
		return l
	case '$':
		var b BakedLabel
		b.Parse(x)
		return b
	case '@':
		var d DelveLabel
		d.Parse(x)
		return d
	default:
		return nil
	}
}

type CfgEntry interface {
	isCfgEntry() struct{}
	fmt.Stringer
}

type ConfiguredLabel struct {
	Name Label
	Cfg  Cfg
}

func (c *ConfiguredLabel) Parse(x *string) {
	c.Name = ParseLabel(x)
	for strings.HasPrefix(*x, "//") {
		*x = strings.TrimPrefix(*x, "//")
	}
	c.Cfg = Cfg{}
	// if strings.HasPrefix(*x, "[") {
	c.Cfg.Parse(x)
	// }
}

func ConfigureLabel(l Label) ConfiguredLabel {
	return ConfiguredLabel{Name: l, Cfg: Cfg{}}
}

func (c ConfiguredLabel) isCfgEntry() struct{} {
	return struct{}{}
}

func (c ConfiguredLabel) String() string {
	return fmt.Sprintf("%s%s", c.Name.String(), c.Cfg.String())
}

type Cfg hashmap.HashMap[ConfiguredLabel, CfgEntry]

func (c Cfg) isCfgEntry() struct{} {
	return struct{}{}
}
func (c Cfg) String() string {
	if len(c) == 0 {
		return ""
	}
	x := "["
	for _, v := range c {
		x = fmt.Sprintf("%s%s=%s,", x, v.Key, v.Value)
	}
	x += "]"
	return x
}
func (c Cfg) Parse(x *string) {
	for strings.HasPrefix(*x, "[") {
		*x = strings.TrimPrefix(*x, "[")
		for !strings.HasPrefix(*x, "]") {
			var cl ConfiguredLabel
			cl.Parse(x)
			*x = strings.TrimPrefix(*x, "=")
			y := ParseCfgEntry(x)
			hashmap.HashMap[ConfiguredLabel, CfgEntry](c).Put(cl, y)
			for strings.HasPrefix(*x, "//") {
				*x = strings.TrimPrefix(*x, "//")
			}
			*x = strings.TrimPrefix(*x, ",")
		}
		*x = strings.TrimPrefix(*x, "]")
	}
}

func ParseCfgEntry(x *string) CfgEntry {
	switch (*x)[0] {
	case 'Q':
		var c ConfiguredLabel
		c.Parse(x)
		return c
	case '$':
		var c ConfiguredLabel
		c.Parse(x)
		return c
	case '@':
		var c ConfiguredLabel
		c.Parse(x)
		return c
	case '[':
		c := Cfg{}
		c.Parse(x)
		return c
	default:
		return nil
	}
}

func init() {
	gob.Register(IpfsLabel(""))
	gob.Register(BakedLabel(""))
	gob.Register(DelveLabel{})
	gob.Register(ConfiguredLabel{})
	gob.Register(Cfg{})
}

/*
 * Copyright (c) 2015-2021 Alex Yatskov <alex@foosoft.net>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of
 * this software and associated documentation files (the "Software"), to deal in
 * the Software without restriction, including without limitation the rights to
 * use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
 * the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package main

import (
	"bytes"
	"fmt"
	"log"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/russross/blackfriday"
)

const (
	defNumCols = 80
	defTabSize = 4
)

const (
	flagNoToc = 1 << iota
	flagNoRules
	flagPascal
)

type list struct {
	index int
}

type heading struct {
	text  []byte
	level int
}

type vimDoc struct {
	filename string
	title    string
	desc     string
	cols     int
	tabs     int
	flags    int
	tocPos   int
	lists    []*list
	headings []*heading
}

func VimDocRenderer(filename, desc string, cols, tabs, flags int) blackfriday.Renderer {
	filename = path.Base(filename)
	title := filename

	if index := strings.LastIndex(filename, "."); index > -1 {
		title = filename[:index]
		if flags&flagPascal == 0 {
			title = strings.ToLower(title)
		}
	}

	return &vimDoc{
		filename: filename,
		title:    title,
		desc:     desc,
		cols:     cols,
		tabs:     tabs,
		flags:    flags,
		tocPos:   -1}
}

func (v *vimDoc) fixupCodeTags(input []byte) []byte {
	r := regexp.MustCompile(`(?m)^\s*([<>])$`)
	return r.ReplaceAll(input, []byte("$1"))
}

func (v *vimDoc) buildHelpTag(text []byte) []byte {
	if v.flags&flagPascal == 0 {
		text = bytes.ToLower(text)
		text = bytes.Replace(text, []byte{' '}, []byte{'_'}, -1)
	} else {
		text = bytes.Title(text)
		text = bytes.Replace(text, []byte{' '}, []byte{}, -1)
	}

	return []byte(fmt.Sprintf("%s-%s", v.title, text))
}

func (v *vimDoc) buildChapters(h *heading) []byte {
	index := -1
	{
		for i, curr := range v.headings {
			if curr == h {
				index = i
				break
			}
		}

		if index < 0 {
			log.Fatal("heading not found")
		}
	}

	var chapters []int
	{
		level := h.level
		siblings := 1

		for i := index - 1; i >= 0; i-- {
			curr := v.headings[i]

			if curr.level == level {
				siblings++
			} else if curr.level < level {
				chapters = append(chapters, siblings)
				level = curr.level
				siblings = 1
			}
		}

		chapters = append(chapters, siblings)
	}

	var out bytes.Buffer
	for i := len(chapters) - 1; i >= 0; i-- {
		out.WriteString(strconv.Itoa(chapters[i]))
		out.WriteString(".")
	}

	return out.Bytes()
}

func (v *vimDoc) writeSplitText(out *bytes.Buffer, left, right []byte, repeat string, trim int) {
	padding := v.cols - (len(left) + len(right)) + trim
	if padding <= 0 {
		padding = 1
	}

	out.Write(left)
	out.WriteString(strings.Repeat(repeat, padding))
	out.Write(right)
	out.WriteString("\n")
}

func (v *vimDoc) writeRule(out *bytes.Buffer, repeat string) {
	out.WriteString(strings.Repeat(repeat, v.cols))
	out.WriteString("\n")
}

func (v *vimDoc) writeToc(out *bytes.Buffer) {
	for _, h := range v.headings {
		title := fmt.Sprintf("%s%s %s", strings.Repeat(" ", (h.level-1)*v.tabs), v.buildChapters(h), h.text)
		link := fmt.Sprintf("|%s|", v.buildHelpTag(h.text))
		v.writeSplitText(out, []byte(title), []byte(link), ".", 2)
	}
}

func (v *vimDoc) writeIndent(out *bytes.Buffer, text string, trim int) {
	lines := strings.Split(text, "\n")

	for index, line := range lines {
		width := v.tabs
		if width >= trim && index == 0 {
			width -= trim
		}

		if len(line) > 0 {
			out.WriteString(strings.Repeat(" ", width))
			out.WriteString(line)
			out.WriteString("\n")
		}
	}
}

// Block-level callbacks
func (v *vimDoc) BlockCode(out *bytes.Buffer, text []byte, lang string) {
	out.WriteString(">\n")
	v.writeIndent(out, string(text), 0)
	out.WriteString("<\n\n")
}

func (v *vimDoc) BlockQuote(out *bytes.Buffer, text []byte) {
	out.WriteString(">\n")
	v.writeIndent(out, string(text), 0)
	out.WriteString("<\n\n")
}

func (v *vimDoc) BlockHtml(out *bytes.Buffer, text []byte) {
	out.WriteString(">\n")
	v.writeIndent(out, string(text), 0)
	out.WriteString("<\n\n")
}

func (v *vimDoc) Header(out *bytes.Buffer, text func() bool, level int, id string) {
	initPos := out.Len()
	if v.flags&flagNoRules == 0 {
		switch level {
		case 1:
			v.writeRule(out, "=")
		case 2:
			v.writeRule(out, "-")
		}
	}

	headingPos := out.Len()
	if !text() {
		out.Truncate(initPos)
		return
	}

	var temp []byte
	temp = append(temp, out.Bytes()[headingPos:]...)
	out.Truncate(headingPos)

	h := &heading{temp, level}
	v.headings = append(v.headings, h)

	tag := fmt.Sprintf("*%s*", v.buildHelpTag(h.text))
	v.writeSplitText(out, bytes.ToUpper(h.text), []byte(tag), " ", 2)
	out.WriteString("\n")
}

func (v *vimDoc) HRule(out *bytes.Buffer) {
	v.writeRule(out, "-")
}

func (v *vimDoc) List(out *bytes.Buffer, text func() bool, flags int) {
	v.lists = append(v.lists, &list{1})
	text()
	v.lists = v.lists[:len(v.lists)-1]
}

func (v *vimDoc) ListItem(out *bytes.Buffer, text []byte, flags int) {
	marker := out.Len()

	list := v.lists[len(v.lists)-1]
	if flags&blackfriday.LIST_TYPE_ORDERED == blackfriday.LIST_TYPE_ORDERED {
		out.WriteString(fmt.Sprintf("%d. ", list.index))
		list.index++
	} else {
		out.WriteString("* ")
	}

	v.writeIndent(out, string(text), out.Len()-marker)

	if flags&blackfriday.LIST_ITEM_END_OF_LIST != 0 {
		out.WriteString("\n")
	}
}

func (*vimDoc) Paragraph(out *bytes.Buffer, text func() bool) {
	marker := out.Len()

	if !text() {
		out.Truncate(marker)
		return
	}

	out.WriteString("\n\n")
}

func (*vimDoc) Table(out *bytes.Buffer, heading []byte, body []byte, columnData []int) {
	// unimplemented
	log.Println("Table is unimplemented")
}

func (*vimDoc) TableRow(out *bytes.Buffer, text []byte) {
	// unimplemented
	log.Println("TableRow is unimplemented")
}

func (*vimDoc) TableHeaderCell(out *bytes.Buffer, text []byte, flags int) {
	// unimplemented
	log.Println("TableHeaderCell is unimplemented")
}

func (*vimDoc) TableCell(out *bytes.Buffer, text []byte, flags int) {
	// unimplemented
	log.Println("TableCell is unimplemented")
}

func (*vimDoc) Footnotes(out *bytes.Buffer, text func() bool) {
	// unimplemented
	log.Println("Footnotes is unimplemented")
}

func (*vimDoc) FootnoteItem(out *bytes.Buffer, name, text []byte, flags int) {
	// unimplemented
	log.Println("FootnoteItem is unimplemented")
}

func (*vimDoc) TitleBlock(out *bytes.Buffer, text []byte) {
	// unimplemented
	log.Println("TitleBlock is unimplemented")
}

// Span-level callbacks
func (*vimDoc) AutoLink(out *bytes.Buffer, link []byte, kind int) {
	out.Write(link)
}

func (*vimDoc) CodeSpan(out *bytes.Buffer, text []byte) {
	out.WriteString("`")
	out.Write(text)
	out.WriteString("`")
}

func (*vimDoc) DoubleEmphasis(out *bytes.Buffer, text []byte) {
	out.Write(text)
}

func (*vimDoc) Emphasis(out *bytes.Buffer, text []byte) {
	out.Write(text)
}

func (*vimDoc) Image(out *bytes.Buffer, link []byte, title []byte, alt []byte) {
	// cannot view images in vim
}

func (*vimDoc) LineBreak(out *bytes.Buffer) {
	out.WriteString("\n")
}

func (*vimDoc) Link(out *bytes.Buffer, link []byte, title []byte, content []byte) {
	out.WriteString(fmt.Sprintf("%s (%s)", content, link))
}

func (*vimDoc) RawHtmlTag(out *bytes.Buffer, tag []byte) {
	// unimplemented
	log.Println("StrikeThrough is unimplemented")
}

func (*vimDoc) TripleEmphasis(out *bytes.Buffer, text []byte) {
	out.Write(text)
}

func (*vimDoc) StrikeThrough(out *bytes.Buffer, text []byte) {
	// unimplemented
	log.Println("StrikeThrough is unimplemented")
}

func (*vimDoc) FootnoteRef(out *bytes.Buffer, ref []byte, id int) {
	// unimplemented
	log.Println("FootnoteRef is unimplemented")
}

// Low-level callbacks
func (v *vimDoc) Entity(out *bytes.Buffer, entity []byte) {
	out.Write(entity)
}

func (v *vimDoc) NormalText(out *bytes.Buffer, text []byte) {
	out.Write(text)
}

// Header and footer
func (v *vimDoc) DocumentHeader(out *bytes.Buffer) {
	if len(v.desc) > 0 {
		v.writeSplitText(out, []byte(v.filename), []byte(v.desc), " ", 0)
	} else {
		out.WriteString(v.filename)
		out.WriteString("\n")
	}

	out.WriteString("\n")
	v.tocPos = out.Len()
}

func (v *vimDoc) DocumentFooter(out *bytes.Buffer) {
	var temp bytes.Buffer

	if v.tocPos > 0 && v.flags&flagNoToc == 0 {
		temp.Write(out.Bytes()[:v.tocPos])

		v.writeRule(&temp, "=")
		title := []byte("Contents")
		tag := fmt.Sprintf("*%s*", v.buildHelpTag(title))
		v.writeSplitText(&temp, bytes.ToUpper(title), []byte(tag), " ", 2)
		temp.WriteString("\n")
		v.writeToc(&temp)
		temp.WriteString("\n")

		temp.Write(out.Bytes()[v.tocPos:])
	} else {
		temp.ReadFrom(out)
	}

	out.Reset()
	out.Write(v.fixupCodeTags(temp.Bytes()))
}

func (v *vimDoc) GetFlags() int {
	return v.flags
}

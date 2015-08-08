/*
 * Copyright (c) 2015 Alex Yatskov <alex@foosoft.net>
 * Author: Alex Yatskov <alex@foosoft.net>
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
	"go/doc"
	"log"
	"strings"

	"github.com/russross/blackfriday"
)

const (
	LIST_STYLE_ORDERED = iota + 1
	LIST_STYLE_UNORDERED
)

func indent(text []byte) []byte {
	return nil
}

type listItem struct {
	style, index int
}

type vimDoc struct {
	lists []*listItem
}

func (v *vimDoc) pushList(style int) {
	v.lists = append(v.lists, &listItem{style, 1})
}

func (v *vimDoc) popList() {
	if len(v.lists) == 0 {
		log.Fatal("invalid list operation")
	}

	v.lists = v.lists[:len(v.lists)-1]
}

func (v *vimDoc) getList() *listItem {
	if len(v.lists) == 0 {
		log.Fatal("invalid list operation")
	}

	return v.lists[len(v.lists)-1]
}

func VimDocRenderer() blackfriday.Renderer {
	return &vimDoc{}
}

func (*vimDoc) hrule(out *bytes.Buffer, repeat string) {
	out.WriteString(strings.Repeat(repeat, 80))
	out.WriteString("\n")
}

func (*vimDoc) formatText(out *bytes.Buffer, text string, level int) {
	doc.ToText(out, string(text), strings.Repeat(" ", 4*level), "", 80)
}

// Block-level callbacks
func (v *vimDoc) BlockCode(out *bytes.Buffer, text []byte, lang string) {
	out.WriteString(">\n")
	v.formatText(out, string(text), 1)
	out.WriteString("<\n\n")
}

func (v *vimDoc) BlockQuote(out *bytes.Buffer, text []byte) {
	out.WriteString(">\n")
	v.formatText(out, string(text), 1)
	out.WriteString("<\n\n")
}

func (v *vimDoc) BlockHtml(out *bytes.Buffer, text []byte) {
	out.WriteString(">\n")
	v.formatText(out, string(text), 1)
	out.WriteString("<\n\n")
}

func (v *vimDoc) Header(out *bytes.Buffer, text func() bool, level int, id string) {
	marker := out.Len()

	switch level {
	case 1:
		v.hrule(out, "=")
	case 2:
		v.hrule(out, "-")
	}

	if !text() {
		out.Truncate(marker)
		return
	}

	out.WriteString(" ~\n\n")
}

func (v *vimDoc) HRule(out *bytes.Buffer) {
	v.hrule(out, "-")
}

func (v *vimDoc) List(out *bytes.Buffer, text func() bool, flags int) {
	style := LIST_STYLE_UNORDERED
	if flags&blackfriday.LIST_TYPE_ORDERED == blackfriday.LIST_TYPE_ORDERED {
		style = LIST_STYLE_ORDERED
	}

	v.pushList(style)
	text()
	v.popList()
}

func (v *vimDoc) ListItem(out *bytes.Buffer, text []byte, flags int) {
	list := v.getList()

	if list.style == LIST_STYLE_ORDERED {
		out.WriteString(fmt.Sprintf("%d. ", list.index))
		list.index++
	} else {
		out.WriteString("* ")
	}

	out.Write(text)

	out.WriteString("\n")
	if flags&blackfriday.LIST_ITEM_END_OF_LIST == blackfriday.LIST_ITEM_END_OF_LIST {
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

func (*vimDoc) Table(out *bytes.Buffer, header []byte, body []byte, columnData []int) {
	// unimplemented
	log.Println("Table is a stub")
}

func (*vimDoc) TableRow(out *bytes.Buffer, text []byte) {
	// unimplemented
	log.Println("TableRow is a stub")
}

func (*vimDoc) TableHeaderCell(out *bytes.Buffer, text []byte, flags int) {
	// unimplemented
	log.Println("TableHeaderCell is a stub")
}

func (*vimDoc) TableCell(out *bytes.Buffer, text []byte, flags int) {
	// unimplemented
	log.Println("TableCell is a stub")
}

func (*vimDoc) Footnotes(out *bytes.Buffer, text func() bool) {
	// unimplemented
	log.Println("Footnotes is a stub")
}

func (*vimDoc) FootnoteItem(out *bytes.Buffer, name, text []byte, flags int) {
	// unimplemented
	log.Println("FootnoteItem is a stub")
}

func (*vimDoc) TitleBlock(out *bytes.Buffer, text []byte) {
	// unimplemented
	log.Println("TitleBlock is a stub")
}

// Span-level callbacks
func (*vimDoc) AutoLink(out *bytes.Buffer, link []byte, kind int) {
	out.Write(link)
}

func (*vimDoc) CodeSpan(out *bytes.Buffer, text []byte) {
	out.Write(text)
}

func (*vimDoc) DoubleEmphasis(out *bytes.Buffer, text []byte) {
	out.Write(text)
}

func (*vimDoc) Emphasis(out *bytes.Buffer, text []byte) {
	out.Write(text)
}

func (*vimDoc) Image(out *bytes.Buffer, link []byte, title []byte, alt []byte) {
	// unimplemented
	log.Println("Image is a stub")
}

func (*vimDoc) LineBreak(out *bytes.Buffer) {
	out.WriteString("\n")
}

func (*vimDoc) Link(out *bytes.Buffer, link []byte, title []byte, content []byte) {
	out.WriteString(fmt.Sprintf("%s (%s)", content, link))
}

func (*vimDoc) RawHtmlTag(out *bytes.Buffer, tag []byte) {
	out.Write(tag)
}

func (*vimDoc) TripleEmphasis(out *bytes.Buffer, text []byte) {
	out.Write(text)
}

func (*vimDoc) StrikeThrough(out *bytes.Buffer, text []byte) {
	// unimplemented
	log.Println("StrikeThrough is a stub")
}

func (*vimDoc) FootnoteRef(out *bytes.Buffer, ref []byte, id int) {
	// unimplemented
	log.Println("FootnoteRef is a stub")
}

// Low-level callbacks
func (v *vimDoc) Entity(out *bytes.Buffer, entity []byte) {
	out.Write(entity)
	// v.formatText(out, string(entity), 0)
}

func (v *vimDoc) NormalText(out *bytes.Buffer, text []byte) {
	out.Write(text)
	// v.formatText(out, string(text), 0)
}

// Header and footer
func (*vimDoc) DocumentHeader(out *bytes.Buffer) {
	// unimplemented
	log.Println("DocumentHeader is a stub")
}

func (*vimDoc) DocumentFooter(out *bytes.Buffer) {
	// unimplemented
	log.Println("DocumentFooter is a stub")
}

func (*vimDoc) GetFlags() int {
	return 0
}

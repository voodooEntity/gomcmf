package converter

import (
	"github.com/voodooEntity/gomcmf/src/util"
	"regexp"
	"strconv"
	"strings"
)

const headingsRxp = `(#+)\s?(.+)`
const imageRxp1 = `!\[([^\]]*)\]\(([^\]]*)\s"(.*)"\)`
const imageRxp2 = `!\[([^\]]*)\]\(([^\]]*)\)`
const linkRxp1 = `\[([^\]]+)\]\(([^\]]*)\s"(.*)"\)`
const linkRxp2 = `\[([^\]]+)\]\(([^\]]*)\)`
const boldRxp1 = `\*\*(.+?)\*\*`
const boldRxp2 = `__(.+?)__`
const italicRxp1 = `\*(.+?)\*`
const italicRxp2 = `_(.+?)_`
const unorderedListItemRxp = `-\s?(.+)`
const blockquoteRxp = `>\s?(.*)`
const codeblockRxp = "```(.*)"

type Content struct {
	Md    string
	Html  string
	State State
}

type State struct {
	IsOpenParagraph   bool
	IsOpenBlock       bool
	InCodeBlock       bool
	InUnorderedList   bool
	IsOpenWrap        bool
	WrapHtml          string
	WrapLinePrefix    string
	EmptyLineCnt      int
	CurrentLine       int
	CurrentLineString string
	LineSplit         []string
}

func (self *Content) Set(content string) {
	self.Md = content
}

func (self *Content) Convert() {
	// open the html and set our converter state
	self.Html = "<div>"
	self.State = State{
		IsOpenParagraph: false,
		IsOpenBlock:     true,
		InCodeBlock:     false,
		InUnorderedList: false,
		EmptyLineCnt:    0,
		LineSplit:       util.Explode("\n", self.Md),
	}

	// now we walk through the whole md document line by line
	// and try to convert it to proper html
	splitText := util.Explode("\n", self.Md)
	for curr, val := range splitText {
		self.State.CurrentLine = curr
		self.State.CurrentLineString = strings.TrimSuffix(val, "\r")
		if !self.State.InCodeBlock {
			if "" == self.State.CurrentLineString {
				self.State.EmptyLineCnt++
			} else {
				// close uls before handling codeblocks, should be handled nicer ###
				self.closeWrap()
				self.handleEmptyLines()
				self.openBlock()
				isHeading := self.handleHeading()
				isListing := self.handleListing()
				isBlockQuote := self.handleBlockQuote()
				isCodeBlock := self.handleCodeBlockOpen()
				if !isHeading && !isListing && !isCodeBlock && !isBlockQuote {
					if !self.openParagraph() {
						self.Html = self.Html + "<br>"
					}
					self.handleSubStringElements()
				}
				self.State.EmptyLineCnt = 0
				self.Html = self.Html + self.State.CurrentLineString
			}
		} else {
			if "```" == self.State.CurrentLineString {
				self.State.InCodeBlock = false
				self.Html = self.Html + "\n    </code></pre>\n"
			} else {
				self.Html = self.Html + "\n" + self.State.CurrentLineString
			}
		}
	}
	self.closeParagraph()
	if self.State.IsOpenBlock {
		self.Html = self.Html + "\n</div>"
	}
}

func (self *Content) openWrap(tag string, prefix string) {
	if !self.State.IsOpenWrap {
		self.State.WrapLinePrefix = prefix
		self.State.WrapHtml = tag
		self.State.IsOpenWrap = true
		self.Html = self.Html + "\n    <" + tag + ">"
	}
}

func (self *Content) closeWrap() {
	if self.State.IsOpenWrap && !strings.HasPrefix(self.State.CurrentLineString, self.State.WrapLinePrefix) {
		self.Html = self.Html + "\n    </" + self.State.WrapHtml + ">"
		self.State.IsOpenWrap = false
	}
}

func (self *Content) closeParagraph() {
	if self.State.IsOpenParagraph {
		self.State.IsOpenParagraph = false
		self.Html = self.Html + "\n  </p>"
	}
}

func (self *Content) openParagraph() bool {
	if !self.State.IsOpenParagraph {
		self.State.IsOpenParagraph = true
		self.Html = self.Html + "\n  <p>\n"
		return self.State.IsOpenParagraph
	}
	return false
}

func (self *Content) openBlock() {
	if !self.State.IsOpenBlock {
		self.State.IsOpenBlock = true
		self.Html = self.Html + "\n<div>"
	}
}

func (self *Content) handleOpenParagraph() {
	if self.State.IsOpenParagraph {
		self.Html = self.Html + "\n  </p>\n"
		self.State.IsOpenParagraph = false
	}
}

func (self *Content) handleEmptyLines() {
	if self.State.EmptyLineCnt == 1 {
		self.handleOpenParagraph()
	} else if 2 <= self.State.EmptyLineCnt {
		self.handleOpenParagraph()
		self.Html = self.Html + "\n</div>"
		self.State.IsOpenBlock = false
	}
	self.State.EmptyLineCnt = 0
}

func (self *Content) handleHeading() bool {
	if !strings.HasPrefix(self.State.CurrentLineString, "#") {
		return false
	}
	self.closeParagraph()
	rxp := regexp.MustCompile(headingsRxp)
	match := rxp.FindStringSubmatch(self.State.CurrentLineString)
	self.State.CurrentLineString = match[2]
	self.handleSubStringElements()
	if self.State.IsOpenParagraph {
		self.State.IsOpenParagraph = false
		self.Html = self.Html + "\n  </p>"
	}
	self.State.CurrentLineString = "\n  <h" + strconv.Itoa(len(match[1])) + ">" + self.State.CurrentLineString + "</h" + strconv.Itoa(len(match[1])) + ">"
	return true
}

func (self *Content) handleSubStringElements() {
	self.handleImages()
	self.handleLinks()
	self.handleBolds()
	self.handleItalics()
}

func (self *Content) handleImages() {
	tmp := regexp.MustCompile(imageRxp1)
	self.State.CurrentLineString = tmp.ReplaceAllString(self.State.CurrentLineString, "      <img src='$2' alt='$1' title='$3'/>")
	tmp = regexp.MustCompile(imageRxp2)
	self.State.CurrentLineString = tmp.ReplaceAllString(self.State.CurrentLineString, "      <img src='$2' alt='$1'/>")
}

func (self *Content) handleCodeBlockOpen() bool {
	if !strings.HasPrefix(self.State.CurrentLineString, "```") {
		return false
	}
	self.closeParagraph()
	self.State.InCodeBlock = true
	tmp := regexp.MustCompile(codeblockRxp)
	self.State.CurrentLineString = tmp.ReplaceAllString(self.State.CurrentLineString, "\n    <pre><code class='language-$1'>")
	return true
}

func (self *Content) handleListing() bool {
	if !strings.HasPrefix(self.State.CurrentLineString, "- ") {
		return false
	}
	//self.openParagraph()
	self.openWrap("ul", "- ")
	rxp := regexp.MustCompile(unorderedListItemRxp)
	match := rxp.FindStringSubmatch(self.State.CurrentLineString)
	self.State.CurrentLineString = match[1]
	self.handleSubStringElements()
	self.State.CurrentLineString = "\n      <li>" + self.State.CurrentLineString + "</li>"
	return true
}

func (self *Content) handleBlockQuote() bool {
	if !strings.HasPrefix(self.State.CurrentLineString, "> ") {
		return false
	}
	//self.openParagraph()
	self.openWrap("blockquote", "> ")
	rxp := regexp.MustCompile(blockquoteRxp)
	match := rxp.FindStringSubmatch(self.State.CurrentLineString)
	self.State.CurrentLineString = match[1]
	self.handleSubStringElements()
	self.State.CurrentLineString = "\n" + self.State.CurrentLineString
	return true
}

func (self *Content) handleLinks() {
	tmp := regexp.MustCompile(linkRxp1)
	self.State.CurrentLineString = tmp.ReplaceAllString(self.State.CurrentLineString, "<a href='$2' title='$3'>$1</a>")
	tmp = regexp.MustCompile(linkRxp2)
	self.State.CurrentLineString = tmp.ReplaceAllString(self.State.CurrentLineString, "<a href='$2'>$1</a>")
}

func (self *Content) handleBolds() {
	tmp := regexp.MustCompile(boldRxp1)
	self.State.CurrentLineString = tmp.ReplaceAllString(self.State.CurrentLineString, "<b>$1</b>")
	tmp = regexp.MustCompile(boldRxp2)
	self.State.CurrentLineString = tmp.ReplaceAllString(self.State.CurrentLineString, "<b>$1</b>")
}

func (self *Content) handleItalics() {
	tmp := regexp.MustCompile(italicRxp1)
	self.State.CurrentLineString = tmp.ReplaceAllString(self.State.CurrentLineString, "<i>$1</i>")
	tmp = regexp.MustCompile(italicRxp2)
	self.State.CurrentLineString = tmp.ReplaceAllString(self.State.CurrentLineString, "<i>$1</i>")
}

package fetch

import (
	"fmt"
	"log"
	"os"

	"../ciweimao"
	"../structure"
	"github.com/cheggaaa/pb"
	"github.com/mitchellh/go-homedir"
)

var txtBar *pb.ProgressBar

func DownloadText(bid string) {
	txtDetail := ciweimao.GetDetail(bid)
	fmt.Println(txtDetail.BookName, "/", txtDetail.AuthorName)
	txtName := txtDetail.BookName
	txtChapters := ciweimao.GetCatalog(bid)
	txtTotalCount := len(txtChapters)
	txtBar = pb.StartNew(txtTotalCount)
	txtContainer := make(map[string]string)
	txtc := make(chan chapterStruct, 409600)
	errc := make(chan structure.ChapterList, 102400)
	txtChaptersArr := splitArray(txtChapters, 16)
	for _, cs := range txtChaptersArr {
		go getChapterText(cs, txtc, errc)
	}
	for {
		select {
		case t := <-txtc:
			txtContainer[t.Cid] = t.Text
			txtBar.Increment()
		case e := <-errc:
			go getChapterText([]structure.ChapterList{e}, txtc, errc)
		}
		if len(txtContainer) == len(txtChapters) {
			close(txtc)
			close(errc)
			break
		}
	}
	txtBar.Finish()
	fmt.Println("writing out files…")
	writeText(txtName, txtContainer, txtChapters)
	fmt.Println("download success!")
}

func writeText(bookName string, txtContainer map[string]string, chapters []structure.ChapterList) {
	bookText := ""
	dir, _ := homedir.Dir()
	expandedDir, _ := homedir.Expand(dir)
	for _, chapter := range chapters {
		bookText += txtContainer[chapter.ChapterID]
	}
	fileName := expandedDir + "/Cirno/download/" + bookName + ".txt"
	if isExist(fileName) {
		os.Remove(fileName)
	}
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println(err.Error())
	}
	_, err = file.WriteString(bookText)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()
}

func getChapterText(chapters []structure.ChapterList, txt chan chapterStruct, errc chan structure.ChapterList) {
	for _, chapter := range chapters {
		text := ""
		chapterInfo, err := ciweimao.GetContent(chapter.ChapterID)
		if err != nil {
			errc <- chapter
		} else {
			text += chapterInfo.ChapterTitle
			text += "\n\n"
			text += chapterInfo.TxtContent
			text += chapterInfo.AuthorSay
			text += "\n\n\n"
			txtstr := chapterStruct{text, chapter.ChapterID}
			txt <- txtstr
		}
	}
}

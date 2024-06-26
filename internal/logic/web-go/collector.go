package web_go

import (
	"fmt"
	"github.com/FirewineXie/envm/util"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	// DefaultURL 提供go版本信息的默认网址
	DefaultURL = "https://golang.google.cn/dl/"
)

// URLUnreachableError URL不可达错误
type URLUnreachableError struct {
	err error
	url string
}

// NewURLUnreachableError 返回URL不可达错误实例
func NewURLUnreachableError(url string, err error) error {
	return &URLUnreachableError{
		err: err,
		url: url,
	}
}

func (e *URLUnreachableError) Error() string {
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("URL %q is unreachable", e.url))
	if e.err != nil {
		buf.WriteString(" ==> " + e.err.Error())
	}
	return buf.String()
}

type Collector struct {
	url string
	doc *goquery.Document
}

// NewCollector 返回采集器实例
func NewCollector(url string) (*Collector, error) {
	if url == "" {
		url = DefaultURL
	}
	c := Collector{
		url: url,
	}
	resp, err := http.Get(c.url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, NewURLUnreachableError(c.url, nil)
	}
	c.doc, err = goquery.NewDocumentFromReader(resp.Body)
	return &c, nil
}

func (c *Collector) loadDocument() (err error) {
	resp, err := http.Get(c.url)
	if err != nil {
		return NewURLUnreachableError(c.url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return NewURLUnreachableError(c.url, nil)
	}
	c.doc, err = goquery.NewDocumentFromReader(resp.Body)
	return err
}

func (c *Collector) findPackages(table *goquery.Selection) (pkgs []*util.Package) {
	alg := strings.TrimSuffix(table.Find("thead").Find("th").Last().Text(), " Checksum")

	table.Find("tr").Not(".first").Each(func(j int, tr *goquery.Selection) {
		td := tr.Find("td")
		pkgs = append(pkgs, &util.Package{
			FileName:  td.Eq(0).Find("a").Text(),
			URL:       td.Eq(0).Find("a").AttrOr("href", ""),
			Kind:      td.Eq(1).Text(),
			OS:        td.Eq(2).Text(),
			Arch:      td.Eq(3).Text(),
			Size:      td.Eq(4).Text(),
			Checksum:  td.Eq(5).Text(),
			Algorithm: alg,
		})
	})
	return pkgs
}

// StableVersions 返回所有稳定版本
func (c *Collector) StableVersions() (items []*VersionGO, err error) {
	c.doc.Find("#stable").NextUntil("#archive").Each(func(i int, div *goquery.Selection) {
		vname, ok := div.Attr("id")
		if !ok {
			return
		}

		versionGO := &VersionGO{}
		versionGO.Name = strings.TrimPrefix(vname, "go")
		versionGO.Packages = c.findPackages(div.Find("table").First())
		items = append(items, versionGO)
	})
	return items, nil
}

// ArchivedVersions 返回已归档版本
func (c *Collector) ArchivedVersions() (items []*VersionGO, err error) {
	c.doc.Find("#archive").Find("div.toggle").Each(func(i int, div *goquery.Selection) {
		vname, ok := div.Attr("id")
		if !ok {
			return
		}
		versionGo := &VersionGO{}
		versionGo.Name = strings.TrimPrefix(vname, "go")
		versionGo.Packages = c.findPackages(div.Find("table").First())
		items = append(items, versionGo)
	})
	return items, nil
}

// AllVersions 返回所有已知版本
func (c *Collector) AllVersions() (items []*VersionGO, err error) {
	items, err = c.StableVersions()
	if err != nil {
		return nil, err
	}
	archives, err := c.ArchivedVersions()
	if err != nil {
		return nil, err
	}
	items = append(items, archives...)
	return items, nil
}

type VersionGO struct {
	util.Version
}

// FindPackage 返回指定操作系统和硬件架构的版本包
func (v *VersionGO) FindPackage(kind, goos, goarch string) (*util.Package, error) {
	if goos == "linux" && goarch == "x86_64" {
		goarch = "386"
	}
	prefix := fmt.Sprintf("go%s.%s-%s", v.Name, goos, goarch)
	for i := range v.Packages {
		if v.Packages[i] == nil || !strings.EqualFold(v.Packages[i].Kind, kind) || !strings.HasPrefix(v.Packages[i].FileName, prefix) {
			continue
		}
		return v.Packages[i], nil
	}
	return nil, util.ErrPackageNotFound
}

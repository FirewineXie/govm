package util

import (
	"crypto/sha1"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

/*
 * @Author: Firewine
 * @File: download
 * @Version: 1.0.0
 * @Date: 2024-04-06 9:31
 * @Description:
 */

// Package 版本安装包
type Package struct {
	FileName    string
	ArchiveName string
	URL         string
	Kind        string
	OS          string
	Arch        string
	Size        string
	Checksum    string
	Algorithm   string // checksum algorithm
}

const (
	// SourceKind go安装包种类-源码
	SourceKind = "Source"
	// ArchiveKind go安装包种类-压缩文件
	ArchiveKind = "Archive"
	// InstallerKind go安装包种类-可安装程序
	InstallerKind = "Installer"
)

// Download 下载版本另存为指定文件并校验sha256哈希值
func (pkg *Package) Download(dst string) (size int64, err error) {
	resp, err := http.Get(pkg.URL)
	if err != nil {
		return 0, NewDownloadError(pkg.URL, err)
	}
	defer resp.Body.Close()
	f, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	size, err = io.Copy(f, resp.Body)
	if err != nil {
		return 0, NewDownloadError(pkg.URL, err)
	}
	return size, nil
}

// DownloadV2 下载版本另存为指定文件并校验sha256哈希值
func (pkg *Package) DownloadV2(dst string) (err error) {
	// Create the file, but give it a tmp file extension, this means we won't overwrite a
	// file until it's downloaded, but we'll remove the tmp extension once downloaded.
	out, err := os.Create(dst + ".tmp")
	if err != nil {
		return err
	}
	resp, err := http.Get(pkg.URL)
	if err != nil {
		return NewDownloadError(pkg.URL, err)
	}
	defer resp.Body.Close()

	parseInt, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	// Create our progress reporter and pass it to be used alongside our writer
	counter := NewOption(0, parseInt)
	_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
	if err != nil {
		return err
	}

	out.Close()

	// The progress use the same line so print a new line once it's finished downloading
	fmt.Print("\n")
	err = os.Rename(dst+".tmp", dst)
	if err != nil {
		return err
	}
	return nil
}

// DownloadError 下载失败错误
type DownloadError struct {
	url string
	err error
}

// NewDownloadError 返回下载失败错误实例
func NewDownloadError(url string, err error) error {
	return &DownloadError{
		url: url,
		err: err,
	}
}

func (e *DownloadError) Error() string {
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("Installation package(%s) download failed", e.url))
	if e.err != nil {
		buf.WriteString(" ==> " + e.err.Error())
	}
	return buf.String()
}

var (
	// ErrUnsupportedChecksumAlgorithm 不支持的校验和算法
	ErrUnsupportedChecksumAlgorithm = errors.New("unsupported checksum algorithm")
	// ErrChecksumNotMatched 校验和不匹配
	ErrChecksumNotMatched = errors.New("file checksum does not match the computed checksum")
)

// VerifyChecksum 验证目标文件的校验和与当前安装包的校验和是否一致
func (pkg *Package) VerifyChecksum(filename string) (err error) {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	var h hash.Hash
	switch pkg.Algorithm {
	case "SHA256":
		h = sha256.New()
	case "SHA1":
		h = sha1.New()
	default:
		return ErrUnsupportedChecksumAlgorithm
	}

	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	if pkg.Checksum != fmt.Sprintf("%x", h.Sum(nil)) {
		return ErrChecksumNotMatched
	}
	return nil
}

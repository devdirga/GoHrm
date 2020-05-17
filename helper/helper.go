package helper

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/creativelab/dbox"
	"github.com/creativelab/knot/knot.v1"
	"github.com/creativelab/toolkit"
)

var (
	DebugMode     bool
	basePath      = (func(dir string, err error) string { return dir }(os.Getwd()))
	StaticImgPath = "/static/img/"
)

var config_system = func() string {
	d, _ := os.Getwd()
	d += "/conf/confsystem.json"
	return d
}()

func GetPathConfig() (result map[string]interface{}) {
	result = make(map[string]interface{})

	ci := &dbox.ConnectionInfo{config_system, "", "", "", nil}
	conn, e := dbox.NewConnection("json", ci)
	if e != nil {
		return
	}

	e = conn.Connect()
	defer conn.Close()
	csr, e := conn.NewQuery().Select("*").Cursor(nil)
	if e != nil {
		return
	}
	defer csr.Close()
	data := []toolkit.M{}
	e = csr.Fetch(&data, 0, false)
	if e != nil {
		return
	}
	result["folder-path"] = data[0].GetString("folder-path")
	result["restore-path"] = data[0].GetString("restore-path")
	result["folder-img"] = data[0].GetString("folder-img")
	return
}

func CreateResult(success bool, data interface{}, message string) map[string]interface{} {
	if !success {
		fmt.Println("ERROR! ", message)
		if DebugMode {
			panic(message)
		}
	}

	return map[string]interface{}{
		"data":    data,
		"success": success,
		"message": message,
	}
}

func UploadHandler(r *knot.WebContext, filename, dstpath string) (error, string) {
	file, handler, err := r.Request.FormFile(filename)
	if err != nil {
		return err, ""
	}
	defer file.Close()

	dstSource := dstpath + toolkit.PathSeparator + handler.Filename
	f, err := os.OpenFile(dstSource, os.O_RDONLY|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err, ""
	}
	defer f.Close()
	io.Copy(f, file)

	return nil, handler.Filename
}

func UploadImage(k *knot.WebContext) (string, error) {
	config := ReadConfig()
	reader, err := k.Request.MultipartReader()
	if err != nil {
		return "", err
	}

	var filename string
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}

		filelocation := filepath.Join(config.GetString("UploadPath"), part.FileName())
		dst, err := os.Create(filelocation)
		if dst != nil {
			defer dst.Close()
		}
		if err != nil {
			return "", err
		}

		filename = part.FileName()
		if _, err := io.Copy(dst, part); err != nil {
			return "", err
		}
	}

	return (StaticImgPath + filename), nil
}

func RenderPathDoc(filename string) string {
	config := ReadConfig()
	filelocation := filepath.Join(config.GetString("DocFilePath"), filename)
	return filelocation
}
func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func Ordinal(x int) string {
	suffix := "th"
	switch x % 10 {
	case 1:
		if x%100 != 11 {
			suffix = "st"
		}
	case 2:
		if x%100 != 12 {
			suffix = "nd"
		}
	case 3:
		if x%100 != 13 {
			suffix = "rd"
		}
	}
	return strconv.Itoa(x) + suffix
}

func ReadConfig() toolkit.M {
	configPath := filepath.Join(basePath, "..", "ecleave-dev", "conf", "newconf.json")
	res := make(toolkit.M)

	bts, err := ioutil.ReadFile(configPath)
	if err != nil {
		toolkit.Println("Error when reading config file.", err.Error())
		os.Exit(0)
	}

	err = toolkit.Unjson(bts, &res)
	if err != nil {
		toolkit.Println("Error when reading config file.", err.Error())
		os.Exit(0)
	}

	return res
}

func GCMEncrypter(text string) string {
	// The key argument should be the AES key, either 16 or 32 bytes
	// to select AES-128 or AES-256.
	key := []byte("AES256Key-32Characters1234567890")
	plaintext := []byte(text)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	ciphertext := nonce
	ciphertext = append(nonce, aesgcm.Seal(nil, nonce, plaintext, nil)...)

	return base64.StdEncoding.EncodeToString(ciphertext)
}

func GCMDecrypter(text string) string {
	// The key argument should be the AES key, either 16 or 32 bytes
	// to select AES-128 or AES-256.
	key := []byte("AES256Key-32Characters1234567890")
	ciphertext, _ := base64.StdEncoding.DecodeString(text)

	nonce := ciphertext[:12]
	ciphertext = ciphertext[12:]

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}

	return string(plaintext)
}
func DistincValue(elements []string) []string {
	// Use map to record duplicates as we find them.
	encountered := map[string]bool{}
	result := []string{}

	for v := range elements {
		if encountered[elements[v]] == true {
			// Do not add duplicate.
		} else {
			// Record this element as an encountered element.
			encountered[elements[v]] = true
			// Append to result slice.
			result = append(result, elements[v])
		}
	}
	// Return the new slice.
	return result
}

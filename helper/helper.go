package helper

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/eaciit/colony-core/v0"
	"github.com/eaciit/dbox"
	_ "github.com/eaciit/dbox/dbc/json"
	_ "github.com/eaciit/dbox/dbc/mongo"
	"github.com/eaciit/knot/knot.v1"
	"github.com/eaciit/sshclient"
	"github.com/eaciit/toolkit"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	DebugMode bool
)

func GetFileExtension(file string) string {
	fileComp := strings.Split(file, ".")
	if len(fileComp) == 0 {
		return ""
	}

	return fileComp[len(fileComp)-1]
}

func Printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Printf(format, a)
}

func HandleError(err error, optionalArgs ...interface{}) bool {
	if err != nil {
		fmt.Printf("error occured: %s", err.Error())

		if len(optionalArgs) > 0 {
			optionalArgs[0].(func(bool))(false)
		}

		return false
	}

	if len(optionalArgs) > 0 {
		optionalArgs[0].(func(bool))(true)
	}

	return true
}

func LoadConfig(pathJson string) (dbox.IConnection, error) {
	connectionInfo := &dbox.ConnectionInfo{pathJson, "", "", "", nil}
	connection, e := dbox.NewConnection("json", connectionInfo)
	if !HandleError(e) {
		return nil, e
	}

	e = connection.Connect()
	if !HandleError(e) {
		return nil, e
	}

	return connection, nil
}

func Connect() (dbox.IConnection, error) {
	connectionInfo := &dbox.ConnectionInfo{"localhost", "eccolonymanager", "", "", nil}
	connection, e := dbox.NewConnection("mongo", connectionInfo)
	if !HandleError(e) {
		return nil, e
	}

	e = connection.Connect()
	if !HandleError(e) {
		return nil, e
	}

	return connection, nil
}

func Recursiver(data []interface{}, sub func(interface{}) []interface{}, callback func(interface{})) {
	for _, each := range data {
		recursiveContent := sub(each)

		if len(recursiveContent) > 0 {
			Recursiver(recursiveContent, sub, callback)
		}

		callback(each)
	}
}

func FetchJSON(url string) ([]map[string]interface{}, error) {
	response, err := http.Get(url)
	if !HandleError(err) {
		return nil, err
	}
	defer response.Body.Close()

	decoder := json.NewDecoder(response.Body)
	data := []map[string]interface{}{}
	err = decoder.Decode(&data)
	if !HandleError(err) {
		return nil, err
	}

	return data, nil
}

func FakeWebContext() *knot.WebContext {
	return &knot.WebContext{Config: &knot.ResponseConfig{}}
}

func RandomIDWithPrefix(prefix string) string {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s%d", prefix, timestamp)
}

func FetchDataSource(_id string, dsType string, path string) ([]map[string]interface{}, error) {
	if dsType == "file" {
		v, _ := os.Getwd()
		filename := fmt.Sprintf("%s/config/datasource/%s", v, path)
		content, err := ioutil.ReadFile(filename)
		if !HandleError(err) {
			return []map[string]interface{}{}, err
		}

		data := []map[string]interface{}{}
		err = json.Unmarshal(content, &data)
		if !HandleError(err) {
			return []map[string]interface{}{}, err
		}

		return data, nil
	} else if dsType == "url" {
		data, err := FetchJSON(path)
		if !HandleError(err) {
			return []map[string]interface{}{}, err
		}

		return data, nil
	}

	return []map[string]interface{}{}, nil
}

func FetchThenSaveFile(r *http.Request, sourceFileName string, destinationFileName string) (multipart.File, *multipart.FileHeader, error) {
	file, handler, err := r.FormFile(sourceFileName)
	if !HandleError(err) {
		return nil, nil, err
	}
	defer file.Close()

	f, err := os.OpenFile(destinationFileName, os.O_WRONLY|os.O_CREATE, 0666)
	if !HandleError(err) {
		return nil, nil, err
	}
	defer f.Close()
	io.Copy(f, file)

	return file, handler, nil
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

func FetchQuerySelector(data []map[string]interface{}, payload map[string]interface{}) ([]map[string]interface{}, error) {
	dataNew := []map[string]interface{}{}
	if len(payload["item"].([]interface{})) > 0 {
		for _, subRaw := range payload["item"].([]interface{}) {
			dataLoop := data
			if len(dataNew) > 0 {
				dataLoop = dataNew
			}
			dataNew = []map[string]interface{}{}
			for _, subData := range dataLoop {
				sub := subRaw.(map[string]interface{})
				tempData := ""
				searchVal := strings.ToLower(sub["name"].(string))
				switch vv := subData[sub["field"].(string)].(type) {
				case string:
					tempData = strings.ToLower(vv)
				case int:
					tempData = strings.ToLower(strconv.Itoa(vv))
				}
				pattern := strings.Index(searchVal, "*")
				if searchVal[:1] == "!" {
					if tempData != searchVal[1:len(searchVal)] {
						dataNew = append(dataNew, subData)
					}
				} else if pattern >= 0 {
					if pattern == 0 {
						searchPat := searchVal[1:len(searchVal)]
						if strings.Index(searchPat, "*") == (len(searchPat) - 1) {
							if strings.Index(tempData, searchPat[0:(len(searchPat)-1)]) >= 0 {
								dataNew = append(dataNew, subData)
							}
						} else {
							if tempData[(len(tempData)-len(searchPat)):len(tempData)] == searchPat {
								dataNew = append(dataNew, subData)
							}
						}
					} else if pattern == (len(searchVal) - 1) {
						if tempData[0:len(searchVal[0:(len(searchVal)-1)])] == searchVal[0:(len(searchVal)-1)] {
							dataNew = append(dataNew, subData)
						}
					}
				} else {
					if tempData == searchVal {
						dataNew = append(dataNew, subData)
					}
				}
			}
		}
	} else {
		dataNew = data
	}
	return dataNew, nil
}

func ToUpper(src string) string {
	regex, err := regexp.Compile("/([A-Z])/g")
	if err != nil {
		fmt.Println(err.Error())
		return src
	}

	return regex.ReplaceAllStringFunc(src, func(w string) string {
		return strings.ToUpper(w)
	})

	// var re = regexp.MustCompile(`\b(` + strings.Join(keywords, "|") + `)\b`)
	// return re.ReplaceAllStringFunc(src, func(w string) string {
	// 	return strings.ToUpper(w)
	// })
}

func GetBetterType(src interface{}) (string, interface{}) {
	if str, ok := src.(float64); ok {
		return "double", str
	} else if str, ok := src.(int64); ok {
		return "int", str
	} else if str, ok := src.(string); ok {
		var value interface{} = nil
		var err error

		value, err = strconv.ParseFloat(str, 64)
		if err == nil {
			if strings.Contains(str, ".") {
				return "double", value
			}

			value, err = strconv.Atoi(str)
			if err == nil {
				return "int", value
			}
		}

		if strings.Contains(str, "ObjectId") {
			value = strings.Replace(strings.Replace(strings.Replace(str, "\"", "", -1), ")", "", -1), "ObjectId(", "", -1)
			return "ObjectId", value
		}

		return "string", str
	} else {
		return "string", src
	}
}

func ForceAsString(data toolkit.M, which string) string {
	return strings.Replace(fmt.Sprintf("%v", data[which]), "<nil>", "", -1)
}

func UploadHandler(r *knot.WebContext, filename, dstpath string) (error, string) {
	file, handler, err := r.Request.FormFile(filename)
	if err != nil {
		return err, ""
	}
	defer file.Close()

	dstSource := dstpath + toolkit.PathSeparator + handler.Filename
	f, err := os.OpenFile(dstSource, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err, ""
	}
	defer f.Close()
	io.Copy(f, file)

	return nil, handler.Filename
}

func ConstructPermission(strPermission string) (result string, err error) {
	permission_map := map[string]string{
		"rwx": "7",
		"rw-": "6",
		"r-x": "5",
		"r--": "4",
		"-wx": "3",
		"-w-": "2",
		"--x": "1",
		"---": "0",
	}

	if len(strPermission) == 9 {
		owner := permission_map[strPermission[:3]]
		group := permission_map[strPermission[3:6]]
		other := permission_map[strPermission[6:]]

		result := owner + group + other
		return result, nil
	} else {

		err := errors.New("The permission is not valid")
		return "", err
	}

	return
}

func ReplaceHostAlias(path string, server colonycore.Server) string {
	for _, alias := range server.HostAlias {
		if strings.Contains(path, alias.HostName) {
			path = strings.Replace(path, alias.HostName, alias.IP, -1)
			break
		}
	}
	return path
}

func RunCommandWithTimeout(sshSetting *sshclient.SshSetting, cmd string, timeout int) error {
	cRunCommand := make(chan string, 1)

	go func() {
		s := *sshSetting
		res, err := s.RunCommandSsh(cmd)
		fmt.Println("cmd    ->", cmd, "output ->", res)

		if err != nil {
			cRunCommand <- err.Error()
		} else {
			cRunCommand <- ""
		}
	}()

	errorMessage := ""
	select {
	case receiveRunCommandOutput := <-cRunCommand:
		errorMessage = receiveRunCommandOutput
	case <-time.After(time.Second * time.Duration(timeout)):
		errorMessage = ""
	}

	if strings.TrimSpace(errorMessage) != "" {
		return errors.New(errorMessage)
	}

	return nil
}

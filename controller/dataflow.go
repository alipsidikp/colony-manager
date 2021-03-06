package controller

import (
	// "github.com/eaciit/controller/data-flow"
	"fmt"
	"strings"
	"time"

	"github.com/eaciit/cast"
	"github.com/eaciit/colony-core/v0"
	"github.com/eaciit/colony-manager/controller/data-flow"
	"github.com/eaciit/colony-manager/helper"
	"github.com/eaciit/dbox"
	"github.com/eaciit/knot/knot.v1"
	tk "github.com/eaciit/toolkit"
)

type DataFlowController struct {
	App
}

func CreateDataFlowController(s *knot.Server) *DataFlowController {
	var controller = new(DataFlowController)
	controller.Server = s
	return controller
}

func Start(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson

	// dataf.Start("test")

	return helper.CreateResult(true, nil, "")
}

func (a *DataFlowController) Save(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson

	payload := map[string]interface{}{}
	err := r.GetPayload(&payload)
	if err != nil {
		return helper.CreateResult(false, nil, err.Error())
	}

	dataShapes := payload["DataShapes"].(map[string]interface{})

	currentDataFlow := new(colonycore.DataFlow)
	currentDataFlow.DataShapes = dataShapes
	currentDataFlow.Actions = constructActions(dataShapes)
	currentDataFlow.Name = tk.ToString(payload["Name"])
	currentDataFlow.Description = tk.ToString(payload["Description"])
	currentDataFlow.ID = tk.ToString(payload["ID"])
	// currentDataFlow.GlobalParam = nil

	dataDs := []colonycore.DataFlow{}
	cursor, err := colonycore.Find(new(colonycore.DataFlow), dbox.Eq("_id", currentDataFlow.ID))
	if cursor != nil {
		cursor.Fetch(&dataDs, 0, false)
		defer cursor.Close()
	}
	if err != nil && cursor != nil {
		return helper.CreateResult(false, nil, err.Error())
	}

	if len(dataDs) == 0 {
		currentDataFlow.CreatedDate = time.Now()
		currentDataFlow.CreatedBy = "Test User"
		currentDataFlow.ID = strings.Replace(currentDataFlow.Name, " ", "", -1) + cast.Date2String(time.Now(), "YYYYMMddHHmm")
	} else {
		currentDataFlow.CreatedDate = dataDs[0].CreatedDate
		currentDataFlow.CreatedBy = dataDs[0].CreatedBy
	}

	currentDataFlow.LastModified = time.Now()

	err = colonycore.Save(currentDataFlow)
	fmt.Println("")
	if err != nil {
		return helper.CreateResult(false, nil, err.Error())
	}
	return helper.CreateResult(true, nil, "success")
}

func constructActions(dataShapes map[string]interface{}) (actions []colonycore.FlowAction) {
	shapes := dataShapes["shapes"]
	connections := dataShapes["connections"].([]interface{})

	conMap := tk.M{}
	for _, val := range connections {
		action := val.(map[string]interface{})
		conMap.Set(action["fromId"].(string), action["toId"].(string))
	}

	for _, val := range shapes.([]interface{}) {
		shape := val.(map[string]interface{})

		dataItem := shape["dataItem"].(map[string]interface{})
		id := shape["id"].(string)
		// firstAction := shape["firstAction"].(bool)
		name := dataItem["name"].(string)

		if name != "" && dataItem["DataAction"] != nil && dataItem["DataActionDetails"] != nil {
			dataAction := dataItem["DataAction"].(map[string]interface{})
			dataActionDetails := dataItem["DataActionDetails"].(map[string]interface{})

			split := strings.Split(dataActionDetails["whenFailed"].(string), "-")

			KO := append([]string{}, strings.Trim(split[1], " "))
			OK := append([]string{}, conMap.GetString(id))

			inputParam := tk.M{}
			outputParam := tk.M{}
			outputType := ""

			if dataActionDetails["output"].(map[string]interface{})["type"] != nil {
				outputType = dataActionDetails["output"].(map[string]interface{})["type"].(string)
			}

			for _, in := range dataActionDetails["input"].([]interface{}) {
				fmt.Printf("=> %#v \n", in)
				inputParam.Set("", "")
			}

			for _, out := range dataActionDetails["output"].(map[string]interface{})["param"].([]interface{}) {
				fmt.Printf("=> %#v \n", out)
				outputParam.Set("", "")
			}

			server := colonycore.Server{}

			if dataAction["server"] != nil {
				dsServer := []colonycore.Server{}
				cursor, err := colonycore.Find(new(colonycore.Server), dbox.Eq("_id", dataAction["server"].(string)))

				if err != nil {
					//
				} else {
					if cursor != nil {
						cursor.Fetch(&dsServer, 0, false)
						defer cursor.Close()
					}

					if len(dsServer) > 0 {
						server = dsServer[0]
					}
				}
			}

			action := colonycore.FlowAction{
				Id:       id,
				Name:     name + "-" + id,
				Interval: 1,
				Retry:    3,
				// FirstAction: firstAction,
				Server:      server,
				InputParam:  inputParam,
				OutputParam: outputParam,
				OutputType:  outputType,
				//   OutputPath: ,
				KO: KO,
				OK: OK,
			}

			switch name {
			case "Spark":
				spark := colonycore.ActionSpark{
					Master:    dataAction["master"].(string),
					Mode:      dataAction["mode"].(string),
					File:      dataAction["appfiles"].(string),
					MainClass: dataAction["mainclass"].(string),
					// Args:      dataAction["args"].(string),
				}
				action.Action = spark
				action.Type = dataflow.ACTION_TYPE_SPARK
				break
			case "Hive":
				hive := colonycore.ActionHive{
					ScriptPath: dataAction["scriptpath"].(string),
					// Params:     ,
				}
				action.Action = hive
				action.Type = dataflow.ACTION_TYPE_HIVE
				break
			case "SSH Script":
				ssh := colonycore.ActionSSH{
					Command: dataAction["script"].(string),
				}
				action.Action = ssh
				action.Type = dataflow.ACTION_TYPE_SSH
				break
			case "HDFS":
				hdfs := colonycore.ActionHDFS{
				// Command: dataAction["script"].(string),
				}
				action.Action = hdfs
				action.Type = dataflow.ACTION_TYPE_HDFS
				break
			case "Java App":
				java := colonycore.ActionJavaApp{
					Jar: dataAction["jar"].(string),
				}
				action.Action = java
				action.Type = dataflow.ACTION_TYPE_JAVA
				break
			case "Map Reduce":
				mr := colonycore.ActionHadoopStreaming{
					// Jar: ,
					Mapper:  dataAction["mapper"].(string),
					Reducer: dataAction["reducer"].(string),
					Input:   dataAction["input"].(string),
					Output:  dataAction["output"].(string),
				}
				action.Action = mr
				action.Type = dataflow.ACTION_TYPE_MAP_REDUCE
				break
			case "Kafka":
				kafka := colonycore.ActionKafka{}
				action.Action = kafka
				action.Type = dataflow.ACTION_TYPE_KAFKA
				break
			case "Email":
				email := colonycore.ActionEmail{
				/*To:,
				  CC:,
				  Subject:,
				  Body:,*/
				}
				action.Action = email
				action.Type = dataflow.ACTION_TYPE_EMAIL
				break
			case "Fork":
				fork := colonycore.ActionFork{}
				action.Action = fork
				action.Type = dataflow.ACTION_TYPE_FORK
				break
			case "Stop":
				stop := colonycore.ActionStop{}
				action.Action = stop
				action.Type = dataflow.ACTION_TYPE_STOP
				break
			}

			actions = append(actions, action)
		}
	}
	return
}

func (a *DataFlowController) GetListData(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson

	payload := map[string]interface{}{}
	err := r.GetPayload(&payload)
	if err != nil {
		return helper.CreateResult(false, nil, err.Error())
	}

	search := tk.ToString(payload["search"])
	var query *dbox.Filter
	if search != "" {
		query = dbox.Or(dbox.Contains("name", search), dbox.Contains("description", search), dbox.Contains("createdby", search))
	}

	cursor, err := colonycore.Find(new(colonycore.DataFlow), query)

	dataDs := []colonycore.DataFlow{}

	if cursor != nil {
		cursor.Fetch(&dataDs, 0, false)
		defer cursor.Close()
	}

	if err != nil && cursor != nil {
		return helper.CreateResult(false, nil, err.Error())
	}

	return helper.CreateResult(true, dataDs, "success")
}

func (a *DataFlowController) Delete(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson

	payload := map[string]interface{}{}
	err := r.GetPayload(&payload)
	if err != nil {
		return helper.CreateResult(false, nil, err.Error())
	}

	ID := tk.ToString(payload["ID"])

	currentDF := new(colonycore.DataFlow)
	err = colonycore.Get(currentDF, ID)
	if err != nil {
		return helper.CreateResult(false, nil, err.Error())
	}

	err = colonycore.Delete(currentDF)
	if err != nil {
		return helper.CreateResult(false, nil, err.Error())
	}

	return helper.CreateResult(true, nil, "success")
}

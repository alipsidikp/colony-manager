package controller

import (
	"github.com/eaciit/colony-core/v0"
	"github.com/eaciit/colony-manager/helper"
	"github.com/eaciit/knot/knot.v1"
	"github.com/eaciit/toolkit"
	"io/ioutil"
	"path/filepath"
)

type PageController struct {
	App
}

func CreatePageController(s *knot.Server) *PageController {
	var controller = new(PageController)
	controller.Server = s
	return controller
}

func (p *PageController) FetchDataSource(ids []string) (toolkit.Ms, error) {
	widgetData := toolkit.Ms{}
	for _, _id := range ids {
		data, err := helper.FetchDataFromDS(_id, 0)
		if err != nil {
			return nil, err
		}
		datasourcewidget := toolkit.M{}
		datasourcewidget.Set("Data", data)
		widgetData = append(widgetData, datasourcewidget)
	}

	return widgetData, nil
}

func (p *PageController) GetAllFields(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson
	type DSMap struct {
		ID     string   `json:"_id"`
		Fields []string `json:"fields"`
	}
	payload := toolkit.M{}
	if err := r.GetPayload(&payload); err != nil {
		return helper.CreateResult(false, nil, err.Error())
	}
	var data = []DSMap{}
	var _data = DSMap{}
	for _, ds := range payload.Get("datasource", "").([]interface{}) {
		tm, err := toolkit.ToM(ds)
		if err != nil {
			return helper.CreateResult(false, nil, err.Error())
		}
		_data.ID = tm.Get("dsWidget", "").(string)
		_data.Fields, err = helper.GetFieldsFromDS(tm.Get("dsColony", "").(string))
		if err != nil {
			return helper.CreateResult(false, nil, err.Error())
		}
		data = append(data, _data)
	}
	widgetSource := filepath.Join(EC_DATA_PATH, "widget", payload.Get("widgetId", "").(string))
	getFileIndex, err := colonycore.GetWidgetPath(widgetSource)
	if err != nil {
		return helper.CreateResult(false, nil, err.Error())
	}
	widgetPath := filepath.Join(getFileIndex, "config-widget.html")
	content, err := ioutil.ReadFile(widgetPath)
	if err != nil {
		return helper.CreateResult(false, nil, err.Error())
	}
	contentstring := string(content)
	/*get widget data*/
	dataWidget := new(colonycore.Widget)
	dataWidget.ID = payload.Get("widgetId", "").(string)
	err = dataWidget.GetById()
	if err != nil {
		return helper.CreateResult(false, nil, err.Error())
	}
	newData := toolkit.M{}
	newData.Set("fieldDs", data)
	newData.Set("container", contentstring)
	newData.Set("pageId", payload.Get("pageId", "").(string))
	newData.Set("url", dataWidget.URL)
	return helper.CreateResult(true, newData, "")
}

func (p *PageController) GetDataSource(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson

	data, err := helper.GetDataSourceQuery()
	if err != nil {
		return helper.CreateResult(false, nil, err.Error())
	}

	return helper.CreateResult(true, data, "")
}

func (p *PageController) GetPage(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson

	payload := map[string]string{}
	if err := r.GetPayload(&payload); err != nil {
		return helper.CreateResult(false, nil, err.Error())
	}
	data, err := new(colonycore.MapPage).Get(payload["search"])
	if err != nil {
		return helper.CreateResult(false, nil, err.Error())
	}
	return helper.CreateResult(true, data, "")
}

func (p *PageController) SavePage(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson

	payload := toolkit.M{}
	err := r.GetPayload(&payload)
	if err != nil {
		return helper.CreateResult(false, nil, err.Error())
	}

	if err := new(colonycore.MapPage).Save(payload); err != nil {
		return helper.CreateResult(false, nil, err.Error())
	}

	return helper.CreateResult(true, payload, "")
}

func (p *PageController) EditPage(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson

	data := colonycore.MapPage{}
	if err := r.GetPayload(&data); err != nil {
		return helper.CreateResult(false, nil, err.Error())
	}
	if err := data.GetById(); err != nil {
		return helper.CreateResult(false, nil, err.Error())
	}

	return helper.CreateResult(true, data, "")
}

func (p *PageController) RemovePage(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson

	payload := map[string]interface{}{}
	if err := r.GetPayload(&payload); !helper.HandleError(err) {
		return helper.CreateResult(false, nil, err.Error())
	}
	idArray := payload["_id"].([]interface{})

	for _, id := range idArray {
		o := new(colonycore.MapPage)
		o.ID = id.(string)
		if err := o.Delete(filepath.Join(EC_APP_PATH, "config", "pages")); err != nil {
			return helper.CreateResult(false, nil, err.Error())
		}
	}

	return helper.CreateResult(true, nil, "")
}

func (p *PageController) SaveConfigPage(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson

	payload := toolkit.M{}
	err := r.GetPayload(&payload)
	if err != nil {
		return helper.CreateResult(false, nil, err.Error())
	}
	if err := new(colonycore.Page).Save(payload, true, ""); err != nil {
		return helper.CreateResult(false, nil, err.Error())
	}

	return helper.CreateResult(true, payload, "")
}

func (p *PageController) SaveDesigner(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson

	payload := toolkit.M{}
	err := r.GetPayload(&payload)
	if err != nil {
		return helper.CreateResult(false, nil, err.Error())
	}
	payload.Set("mode", "save widget")
	if err := new(colonycore.Page).Save(payload, true, ""); err != nil {
		return helper.CreateResult(false, nil, err.Error())
	}

	return helper.CreateResult(true, payload, "")
}

func (p *PageController) EditPageDesigner(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson

	payload := toolkit.M{}
	data := colonycore.Page{}
	if err := r.GetPayload(&payload); err != nil {
		return helper.CreateResult(false, nil, err.Error())
	}
	data.ID = payload.Get("_id", "").(string)

	if payload.Get("mode", "").(string) != "" {
		data.Save(payload, true, filepath.Join(EC_DATA_PATH, "widget"))
	}
	if err := data.GetById(); err != nil {
		return helper.CreateResult(false, nil, err.Error())
	}

	return helper.CreateResult(true, data, "")
}

/*func (p *PageController) PreviewExample(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson

	data := toolkit.M{}
	if err := r.GetPayload(&data); err != nil {
		return helper.CreateResult(false, nil, err.Error())
	}
	widgetSource := filepath.Join(EC_DATA_PATH, "widget", data.Get("_id", "").(string))

	getFileIndex, err := colonycore.GetPath(widgetSource)
	if err != nil {
		return helper.CreateResult(false, nil, err.Error())
	}
	widgetPath := filepath.Join(getFileIndex, "index.html")

	content, err := ioutil.ReadFile(widgetPath)
	if err != nil {
		return helper.CreateResult(false, nil, err.Error())
	}
	contentstring := string(content)

	var datasource []string
	for _, val := range data.Get("dataSource").([]interface{}) {
		datasource = append(datasource, val.(string))
	}

	dataSourceArry := strings.Join(datasource, ",")
	widgetData, err := new(WidgetController).FetchDataSources(dataSourceArry)
	if err != nil {
		return helper.CreateResult(false, nil, err.Error())
	}

	previewData := toolkit.M{}
	previewData.Set("container", contentstring)
	previewData.Set("dataSource", widgetData)

	if data.Get("mode", "").(string) == "save" {
		dataWidget := colonycore.Widget{}
		dataWidget.ID = data.Get("_id", "").(string)
		if err := dataWidget.GetById(); err != nil {
			return helper.CreateResult(false, nil, err.Error())
		}
		dataWidget.DataSourceId = datasource

		configs := toolkit.Ms{}
		for _, val := range data.Get("config", "").([]interface{}) {
			configs = append(configs, val.(map[string]interface{}))
		}
		dataWidget.Config = configs

		if err := dataWidget.Save(); err != nil {
			return helper.CreateResult(false, nil, err.Error())
		}
	}

	return helper.CreateResult(true, previewData, "")
}*/

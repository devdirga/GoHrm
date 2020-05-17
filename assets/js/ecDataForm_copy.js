$.fn.ecDataForm = function (options) {
    var settings = $.extend({
        currentMode: "",
        defaultMode: "",
        gridDataPoint: null,
        formSaveDataPoint: null,
        formGetDataPoint: null,
        dataModel: null,
        formButtons: [],
        gridButtons: [],
        dataFormId: "",
        gridColumns: [],
        dataSources: []
    }, options);

    this.panelGrid = null;
    this.panelForm = null;
    this.toolbarGrid = null;
    this.objectGrid = null;
    this.toolbarForm = null;
    this.objectForm = null;
    this.formModel = null;
    this.dataSources = [];
    this.currentMode = settings.currentMode;

    this.render = function () {
        if (settings.dataFormId == "" || settings.dataFormId == undefined) settings.dataFormId = "";
        var othis = this;
        var $this = $(othis);
        $this.addClass("ec_df");

        othis.dataSources = settings.dataSources;

        //-- create panel for grid
        othis.panelGrid = $("<div id=\"panelGrid_" + settings.dataFormId + "\" class='ec_df_panel'></div>").appendTo($this);

        //-- create toolbar & object for Grid
        othis.toolbarGrid =
            $("<div class='col-md-12'></div>").appendTo(
            $("<div id=\"gridToolbar_" + settings.dataFormId + "\" class='ec_df_toolbar row'></div>").appendTo(othis.panelGrid));
        othis.objectGrid = $("<div id=\"objectGrid_" + settings.dataFormId + "\"></div>").appendTo(othis.panelGrid);

        var buttonAddNew = $("<button class=\"btn btn-sm btn-xs btn-primary\">Add New</button>").appendTo(othis.toolbarGrid);
        var buttonFilterDialog = $("<button class=\"btn btn-sm btn-xs btn-primary\">Show Filter</button>").appendTo(othis.toolbarGrid);
        buttonAddNew.click(function () {
            othis.add();
        });

        //-- create panel for form
        othis.panelForm = $("<div id=\"panelForm_" + settings.dataFormId + "\" class='ec_df_panel form-horizontal'></div>").appendTo($this);
        othis.toolbarForm =
            $("<div class='col-md-12'></div>").appendTo(
            $("<div id=\"gridToolbar_" + settings.dataFormId + "\" class='ec_df_toolbar form-horizontal'></div>").appendTo(othis.panelForm));
        othis.objectForm = $("<div id=\"objectForm_" + settings.dataFormId + "\"></div>").appendTo(othis.panelForm);
        settings.formFields.forEach(function (obj) {
            buildFormField(obj, othis);
            //.appendTo(othis.objectForm);
        });

        $("<button class=\"btn btn-sm btn-xs btn-warning\">Cancel</button>")
            .click(function () {
                othis.changeMode("");
            }).appendTo(this.toolbarForm);
        $("<button class=\"btn btn-sm btn-xs btn-primary\">Save</button>")
            .click(function () {
                //-- save process
                othis.save(function () {
                    othis.changeMode("");
                });
                //-- end of save process
            }).appendTo(this.toolbarForm);

        if (settings.dataModel != null) {
            othis.formModel = kendo.observable(settings.dataModel);
            kendo.bind(this.panelForm, othis.formModel);
        }

        othis.gridDataPoint = new ecDataPoint(settings.gridDataPoint);
        othis.formGetDataPoint = new ecDataPoint(settings.formGetDataPoint);
        othis.formSaveDataPoint = new ecDataPoint(settings.formSaveDataPoint);
        othis.changeMode("");

        //-- kendoit
        var dsGrid = new kendo.data.DataSource({
            data: [],
            pageSize: 10
        });
        othis.objectGrid.kendoGrid({
            dataSource: dsGrid,
            columns: settings.gridColumns,
            sortable: true
        });

        othis.populate();

        othis.data("ecDataForm", othis);
        return othis;
    }

    this.save = function (onSaveOk) {
        var othis = this;
        othis.formSaveDataPoint.run(this.formModel, onSaveOk);
    }

    this.add = function () {
        var othis = this;
        kendo.unbind(this.panelForm);
        othis.formModel = kendo.observable(settings.dataModel);
        kendo.bind(othis.panelForm, othis.formModel);
        othis.changeMode("New");
    }

    this.get = function (onGetOk, onBindingOk) {
        var othis = this;
        othis.formGetDataPoint.run(this.formModel, function (data) {
            kendo.unbind(this.panelForm);
            if (typeof onGetOk == 'function') data = onGetOk(data);
            othis.formModel = kendo.observable(data.Data);
            kendo.bind(othis.panelForm, othis.formModel);
            if (typeof onBindingOk == 'function') data = onBindingOk(data);
        });
    }

    this.populate = function (onPopulateOk, onBindingOk) {
        var othis = this;
        othis.gridDataPoint.run(this.formModel, function (data) {
            if (typeof onPopulateOk == 'function') data = onPopulateOk(data);
            othis.objectGrid.data("kendoGrid").setDataSource(new kendo.data.DataSource({ data: data.Data }));
            if (typeof onBindingOk == 'function') data = onBindingOk(data);
            return data;
        });
    }

    this.changeMode = function (newMode) {
        var othis = this;
        othis.currentMode = newMode;
        switch (othis.currentMode) {
            case "":
                this.panelGrid.show();
                this.panelForm.hide();
                this.populate();
                break;

            case "Edit":
            case "New":
                this.panelGrid.hide();
                this.panelForm.show();
                break;
        }
    }

    return this;
};

function ecdfFieldVisibility(m, all, edit, insert) {
    var v = true
    if (all == false) {
        v = false;
    }
    else {
        v = (m == "Edit" && edit) ||
            (m == "New" && insert);
    }
    return v;
}

function ecdfFieldReadOnly(m, all, edit, insert) {
    var v = true
    if (all == false) {
        v = false;
    }
    else {
        v = (m == "Edit" && edit) ||
            (m == "New" && insert);
    }
    return v;
}

function ecdfShow(labelOnControl, currentMode, v0, v1, v2, r0, r1, r2) {
    var s = true;
    var v = ecdfFieldVisibility(currentMode, v0, v1, v2);
    var r = ecdfFieldReadOnly(currentMode, r0, r1, r2);
    switch (labelOnControl.toLowerCase()) {
        case "control":
            s = currentMode != "" && v && !r;
            break;

        case "text":
            s = currentMode == "" && v && r;
            break;
    }
    return s;
}

function buildFormField(obj, ecdf) {
    var objForm = ecdf.objectForm;
    var dataSources = ecdf.dataSources;
    var controlType = "text";
    var f = "";

    var vi = [gp(obj, "visible", true), gp(obj, "visibleNew", true), gp(obj, "visibleEdit", true)];
    var ro = [gp(obj, "readOnly", true), gp(obj, "readOnlyNew", true), gp(obj, "readOnlyEdit", true)];

    var textInfo = "";
    //"<span data-bind=\"text:" + obj.field + "," +
    //"visible:"+vi[0]+"||(currentMode=='New'&&v[1])||(currentMode=='Edit'&&v[2])\" />";
    //"visible:ecdfContext(this).currentMode=='Edit'\" />";

    if (obj.hasOwnProperty("template")) {
        var t = $("#" + obj.template);
        if (t != undefined) {
            if (objForm != undefined) {
                var d = $("<div></div>").appendTo(objForm);
                d.html(t.html());
                return d;
            };
        }
        else {
            return $("");
        }
    }

    var ds = undefined;
    if (obj.hasOwnProperty("control")) controlType = obj.control;
    if (obj.hasOwnProperty("dataSource") && dataSources != null && dataSources.length > 0) {
        ds = Enumerable.From(dataSources).FirstOrDefault(undefined).run();
    }
    switch (controlType) {
        case "":
        case "text":
            f = $("<div class='form-group'>" +
                    "<label class='col-md-2 align_right'>" + obj.title + "</label>" +
                    "<div class='col-md-10'>" +
                    textInfo +
                    "<input type='text' class='input-sm form-control' " +
                    "placeholder='Please enter " + obj.title + "' data-bind='value:" + obj.field + "'>" +
                    "</div>" +
                    "</div>");
            break;

        case "select":
            var os = "";
            if (typeof ds != undefined) {
                ds.forEach(function (i, v) {
                    if (typeof i == "string" || isNaN(i) == false) {
                        os += "<option value='" + i + "'>" + i + "</option>\n";
                    }
                });
            }
            f = $("<div class='form-group'>" +
                    "<label class='col-md-2 align_right'>" + obj.title + "</label>" +
                    "<div class='col-md-10'>" +
                    textInfo +
                    "<select type='text' placeholder='Please enter " + obj.title + "' data-bind='value:" + obj.field + "'>" + os + "</select>" +
                    "</div>" +
                    "</div>");
            break;
    };
    if (objForm != undefined) f.appendTo(objForm);
    return f;
}

function ecdfGridSelect(obj) {
    var j = $(obj);
    try {
        var _id = j.parents("tr").find("#record_check").val();
        var ecdf = j.parents("div.ec_df").data("ecDataForm");
        ecdf.formModel.set("_id", _id);
        ecdf.get(function (data) {
            ecdf.changeMode("Edit");
            return data;
        });
    }
    catch (e) {
        alert("Invalid call: " + e);
    }
}

function ecdfContext(obj) {
    var j = $(obj);
    var ecdf = j.parents("div.ec_df").data("ecDataForm");
    return ecdf;
}
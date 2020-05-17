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
        modelFunctions: {},
        dataFormId: "",
        gridColumns: [],
        dataSources: [],
        formFields: [],
        initForm: undefined,
        initGrid: undefined
    }, options);

    var ec = this;
    this.panelGrid = null;
    this.panelForm = null;
    this.toolbarGrid = null;
    this.objectGrid = null;
    this.toolbarForm = null;
    this.objectForm = null;
    this.dataSources = [];
    this.model = {
        functions: settings.modelFunctions,
        currentMode: ko.observable(settings.currentMode),
        record: ko.observable({}),
        save: function (onSaveOk) {
            var mthis = ec.model;
            mthis.formSaveDataPoint().run(mthis.record(), function () {
                mthis.currentMode("");
                ec.populate();
                if (typeof onSaveOk == "function") onSaveOk(data);
            });
        },
        add:function(){
            var mthis = ec.model;
            mthis.record(ko.mapping.fromJS(settings.dataModel));
            mthis.currentMode("New");
            ec.initForm();
        },
        get: function (onGetOk, onBindingOk) {
            var mthis = ec.model;
            mthis.formGetDataPoint().run(mthis.record(), function (data) {
                if (typeof onGetOk == 'function') data = onGetOk(data);
                mthis.record(ko.mapping.fromJS(data.Data));
                if (typeof onBindingOk == 'function') data = onBindingOk(data);
                ec.initForm();
            });
        },
        cancel: function () {
            ec.model.currentMode("");
        }
    };
    
    this.render = function () {
        if (settings.dataFormId == "" || settings.dataFormId == undefined) settings.dataFormId = "";
        var othis = this;
        var $this = $(othis);
        $this.addClass("ec_df");
    
        othis.dataSources = settings.dataSources;
        
        //-- create panel for grid
        othis.panelGrid = $("<div id=\"panelGrid_" + settings.dataFormId + "\" class='ec_df_panel' " +
            "data-bind=\"visible:currentMode()==''\"></div>").appendTo($this);

        //-- create toolbar & object for Grid
        othis.toolbarGrid =
            $("<div class='col-md-12'></div>").appendTo(
            $("<div id=\"gridToolbar_" + settings.dataFormId + "\" class='ec_df_toolbar row'></div>").appendTo(othis.panelGrid));
        othis.objectGrid = $("<div id=\"objectGrid_" + settings.dataFormId + "\"></div>").appendTo(othis.panelGrid);
        
        var buttonAddNew = $("<button class=\"btn btn-sm btn-primary\" data-bind=\"click:add\">Add New</button>").appendTo(othis.toolbarGrid);
        var buttonFilterDialog = $("<button class=\"btn btn-sm btn-primary\">Show Filter</button>").appendTo(othis.toolbarGrid);

        //-- create panel for form
        othis.panelForm = $("<div id=\"panelForm_" + settings.dataFormId + "\" " +
            "data-bind=\"visible:currentMode()!='',with:record\" " +
            "class=\"ec_df_panel\"></div>")
            .appendTo($this);
        othis.toolbarForm =
            $("<div class='col-md-12'></div>").appendTo(
            $("<div id=\"gridToolbar_" + settings.dataFormId + "\" class='ec_df_toolbar row'></div>")
            .appendTo(othis.panelForm));
        othis.objectForm = $("<div id=\"objectForm_" + settings.dataFormId + " row\"></div>").appendTo(othis.panelForm);
        settings.formFields.forEach(function (fld) {
            //buildFormGroup(grp, othis);
            new ecFormField(fld).render(othis,othis.objectForm);
        });

        $("<button class=\"btn btn-sm btn-warning\" data-bind='click:$root.cancel'>Cancel</button>").appendTo(this.toolbarForm);
        $("<button class=\"btn btn-sm btn-primary\" data-bind='click:$root.save'>Save</button>")
            .appendTo(this.toolbarForm);

        if (settings.dataModel != null) {
            //othis.model.record() = kendo.observable(settings.dataModel);
            othis.model.record(ko.mapping.fromJS(settings.dataModel));
            ko.cleanNode(othis);
            ko.applyBindings(othis.model, $(othis)[0]);
        }

        othis.gridDataPoint = new ecDataPoint(settings.gridDataPoint);
        othis.model.formGetDataPoint = new ecDataPoint(settings.formGetDataPoint);
        othis.model.formSaveDataPoint = new ecDataPoint(settings.formSaveDataPoint);
        othis.model.currentMode("");

        if (gp(othis, "toolbarButtons", undefined) != undefined) {

        }

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

        if(settings.currentMode=="")othis.populate();
        return othis;
    }

    this.populate = function (onPopulateOk, onBindingOk) {
        var othis = this;
        othis.gridDataPoint.run(this.model.record(), function (data) {
            if (typeof onPopulateOk == 'function') data=onPopulateOk(data);
            othis.objectGrid.data("kendoGrid").setDataSource(new kendo.data.DataSource({data:data.Data}));
            if (typeof onBindingOk == 'function') data=onBindingOk(data);
            return data;
        });
    }

    this.initForm = function () {
        var othis = this;

        //-- manage datepicker
        input2datePicker(othis.find(".ec_df_date"));
        
        $.each(othis.find(".ec-df-treeview"), function (idx, obj) {
            var d = $(obj);
            //alert(d.html());
            var ds = Enumerable.From(othis.dataSources)
                .FirstOrDefault(undefined, "$.id=='" + d.attr("datasource") + "'");
            if (ds != undefined) {
                ds.run(othis.model.record(), function (data) {
                    var dataTreeView = data.Data;
                    var dsTreeView = new kendo.data.HierarchicalDataSource({
                        data: dataTreeView
                    });

                    var tvw = d.data("kendoTreeView");
                    if (tvw == undefined) {
                        tvw = d.kendoTreeView({
                            dataSource: dsTreeView
                        }).data("kendoTreeView");
                    }
                    else {
                        tvw.setDataSource(dsTreeView);
                    }
                    tvw.expand(".k-item");
                });
            }
        });

        //-- run additional init
        if (typeof settings.initForm == 'function') {
            settings.initForm(othis);
        }
    }

    this.data("ecDataForm", this);
    return this;
};

function ecdfFieldVisibility(m, all, edit, insert){
    var v = true
    if(all==false){
        v = false;
    }
    else{
        v = (m == "Edit" && edit) ||
            (m == "New" && insert);
    }
    return v;
}

function ecdfFieldReadOnly(m, all, edit, insert){
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
            s = currentMode!=""  && v && !r;
            break;

        case "text":
            s = currentMode=="" && v && r;
            break;
    }
    return s;
}

function ecdfGridSelect(obj) {
    var j = $(obj);
    try{
        var _id = j.parents("tr").find("#record_check").val();
        var ecdf = j.parents("div.ec_df").data("ecDataForm");
        ecdf.model.record()._id(_id);
        ecdf.model.get(function (data) {
            ecdf.model.currentMode("Edit");
            return data;
        });
    }
    catch (e) {
        alert("Invalid call: "+e);
    }
}

function ecdfContext(obj) {
    var j = $(obj);
    var ecdf;
    if (j.is("div.ec_df")) {
        ecdf = j.data("ecDataForm");
    }
    else {
        ecdf = j.parents("div.ec_df").data("ecDataForm");
    }
    return ecdf;
}
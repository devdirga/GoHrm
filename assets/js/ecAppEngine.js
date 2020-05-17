$.fn.ecAppEngine = function (options) {
    var settings = $.extend({
        dataPoints: {},
        dataModel: null,
        columnSettings: {},
        formFields: {},
        scripts: {},
        controls: {}
    }, options);

    var othis = this;
    othis.dataPoints = settings.dataPoints;
    othis.dataModel = settings.dataModel;
    othis.columnSettings = settings.columnSettings;
    othis.scripts = settings.scripts;
    othis.controls = settings.controls;
    othis.formFields = settings.formFields;
    othis.data("ecAppEngine", othis);
    
    this.render = function () {
        var $this = $(this);
        $.each($("[control='toolbar']"), function (idx, obj) {
            var $obj = $(obj);
            var buttons = new String($obj.attr("buttons"));
            var btn = buttons.split(",");
            btn.forEach(function (objname) {
                if (othis.controls.hasOwnProperty(objname)) {
                    var nc = othis.controls[objname];
                    $obj.append("<button class=\""+nc.class+"\" onclick=\"ecRunScript(this)\" script=\""+gp(nc,"click","")+"\">" + nc.title + "</button>");
                }
            });
        });

        $.each(othis.find("[control='formview']"), function (idx, obj) {
            var formObj = $(obj);
            var fieldId = formObj.attr("formfields");
            if (fieldId != undefined) {
                var fields = gp(othis.formFields, fieldId, []);
                fields.forEach(function (fld) {
                    fld.modeField = formObj.attr("modefield");
                    new ecFormField(fld).render(othis, formObj);
                });
            }
        });

        ko.cleanNode(othis[0]);
        ko.applyBindings(model, othis[0]);
        //alert(othis.dataModel().record());

        var populateScripts = [];
        $.each(othis.find("[control='grid']"), function (idx, obj) {
            ecGridInit(obj, othis, false);
            var gridObj = $(obj);
            var populateScript = gridObj.attr("populate");
            var autoPopulate = gridObj.attr("autopopulate");
            if (populateScript != undefined && autoPopulate=="true" && $.inArray(populateScript, populateScripts)) {
                populateScripts.push(populateScript);
            }
        });
        populateScripts.forEach(function (s) {
            othis.run(s, othis);
        });

        input2datePicker($("input[control='datetime']"));
        othis.data("ecAppEngine", othis);
    }

    this.run = function(scriptName,p1,p2,p3,p4,p5,p6,p7){
        var ectx = othis;
        if (scriptName != undefined
            && ectx.scripts.hasOwnProperty(scriptName)
            && typeof ectx.scripts[scriptName] == 'function'
            ) {
            var fn = ectx.scripts[scriptName];
            fn(p1, p2, p3, p4, p5, p6, p7);
        }
    }
    return this;
}

function ecContext(obj) {
    var ectx = obj.parents("div.ecis-app").data("ecAppEngine");
    return ectx;
}

function ecGridInit(obj, ectx, populate) {
    var gridObj = $(obj);
    var othis = ectx==undefined ? ecContext(gridObj) :  ectx;
    var columnSetting = {
        columns: othis.columnSettings[gridObj.attr("columnsetting")],
        sortable: true,
        filterable: true,
        resizable: true,
        pageable: true
    };
    gridObj.kendoGrid(columnSetting);
    var populateScript = gridObj.attr("populate");
    if (populateScript != undefined && populate == true) {
        othis.run(populateScript, othis);
    }
}

function ecGridSelect(obj) {
    var j = $(obj);
    try {
        var ectx = ecContext(j);
        var _id = j.parents("tr").find("#record_check").val();
        //ectx.dataModel().record({ Title: ko.observable(_id) });
        var g = j.parents("div[control='grid']");
        var fnName = g.attr("select");
        if (fnName != undefined && ectx.scripts.hasOwnProperty(fnName)) {
            var fn = ectx.scripts[fnName];
            if (typeof fn == "function") {
                fn(obj, ectx, _id);
            }
        }
    }
    catch (e) {
        alert("Invalid call: " + e);
    }
}

function ecRunScript(obj) {
    var $obj = $(obj);
    var ectx = ecContext($obj);
    var scriptName = $obj.attr("script");
    if (scriptName != undefined 
        &&  ectx.scripts.hasOwnProperty(scriptName)
        && typeof ectx.scripts[scriptName]=='function'
        ) 
    {
        var fn = ectx.scripts[scriptName];
        fn(obj,ectx);
    }
}

function ecGridPopulate(scriptname, ectx, onPopulateOk, onBindingOk) {
    var objs = $("[control='grid'][populate='"+scriptname+"']");
    $.each(objs, function (i, o) {
        var obj = $(o);
        var dp = ectx.dataPoints[obj.attr("datasource")];
        if (dp != undefined) {
            dp.run(ectx.dataModel(), function (data) {
                if (typeof onPopulateOk == 'function') data = onPopulateOk(data);
                obj.data("kendoGrid")
                    .setDataSource(new kendo.data.DataSource({ data: data.Data, pageSize: 10 }));
                if (typeof onBindingOk == 'function') data = onBindingOk(data);
                return data;
            });
        }
        else {
            return [];
        }
    });
}
            
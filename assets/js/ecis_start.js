function ecisStart()
{
    //kendo.culture(cultureInfo);
    //--- initiate toogle
    $.each($("[ecis_role='checkbox_toogle']"), function (index, obj) {
        alert(obj.attr("ecis_checkbox_toogleid"));
    });

    /*
    $.each($(".entry_date"),function(idx, obj){
        var jobj = $(obj);
        var fmt = jobj.attr("format") == undefined ? jsonDateFormat : jobj.attr("format");
        var depth = jobj.attr("depth") == undefined ? "month" : jobj.attr("depth");
        var start = jobj.attr("start") == undefined ? "month" : jobj.attr("start");
        jobj.css("width", "100px").kendoDatePicker({ format: fmt, start: start, depth: depth });
    });
    //$(".dateSelector").datepicker();
    */
    input2datePicker($("input.entry_date"));
    input2timePicker($("input.entry_time"));
    ecisInitControl($("body"));
}

function multiSelectValue(obj, returnType) {
    if (obj.data("kendoMultiSelect") == undefined) return;
    var value = obj.data("kendoMultiSelect").value();
    var ret = value;
    if (arguments.length > 1 && arguments[1].toLowerCase() == "string") {
        ret = value.join("|");
    }
    return ret;
}

function toArray(objs) {
    return $.map(objs, function (obj, idx) { return obj; });
}

function loadJs(scriptPath) {
    var oScript = $("<scr" + "ipt" + "></scr" + "ipt>");
    oScript.attr("type", "text/javascript").attr("src", scriptPath);
    $("#script_dynamic").append(oScript);
}

function getMenuObj() {
    var menuObj = undefined;
    var menu_type = "";
    var menui_id = "";

    if (arguments.length == 1) {
        menu_type = arguments[0].attr("menu_type");
        module_id = arguments[0].attr("module_id");
    }
    else {
        menu_type = arguments[0];
        module_id = arguments[1];
    }
    menuObj = Enumerable.From(menu_type == "favorite" ? ecisMenus.favorites : ecisMenus.menus)
            .Where("$.module_id=='" + module_id + "'")
            .FirstOrDefault(undefined);
    return menuObj;
}

function devide(v1, v2) {
    var ret = 0;
    if (v2 == 0) {
        ret = 0;
    }
    else {
        ret = v1 / v2;
    }
    return ret;
}

function date2time(time)
{
    return kendo.format("{0:HH:mm}", time);
}

function time2decimal(time, nextDay) {
    var times = time.split(":");
    if (times[0] == "00" && nextDay==1) times[0] = "24";
    return parseFloat(times[0]) + (times[1] / 60.00);
}

function decimal2time(time)
{
    var hours = Math.floor(time * 10 / 10);
    var mins = time - hours;
    return hours.toString() + ":" + mins;
}

function input2timePicker(objects) {
    $.each(objects, function (idx, obj) {
        var jobj = $(obj);
        if (jobj.data("kendoTimePicker") == undefined) {
            jobj.kendoTimePicker({format:"HH:mm", interval:30});
        }
    });
}

function ecisInitControl(ctlObj) {
    if (!ctlObj.is("body"))
    {
        alert("OK");
    }
}

function getObjectProperties(obj) {
    var keys = [];
    for (var key in obj) {
        keys.push(key);
    }
    return keys;
}

function getObjectProperty(obj, prop, def) {
    if (def == undefined) def = "";
    if (obj.hasOwnProperty(prop))
        return obj[prop];
    else return def;
}

function toggleCheckbox(obj, selectorTxt) {
    var cbxs = $(selectorTxt);
    var checked = obj.prop("checked");
    cbxs.prop("checked", checked);
}


function gridDelete(deleteProcessUrl, deletedCheckboxes, fnDeleteSuccess, fnNoDelete) {
    var DeletedIds = $.map(deletedCheckboxes,function(obj,idx){
        return $(obj).val();
    });
    if (DeletedIds.length > 0) {
        if (!confirm("Are you sure you want to delete selected record(s) ?")) return;
        executeOnServer(viewModel, deleteProcessUrl + encodeURIComponent(ko.mapping.toJSON(DeletedIds)), fnDeleteSuccess);
    }
    else {
        if (typeof fnNoDelete == "function") fnNoDelete();
    }
}

function hideDialog(ow) {
    if (ow.data("kendoWindow")) {
        var kw = coalesce(ow.data("kendoWindow"), ow.kendoWindow({modal:false}).data("kendoWindow"));
        kw.close();
    }
}

function showDialog(ow, title, fnClose) {
    var kw = null;
    if (!ow.data("kendoWindow")) {
        ow.kendoWindow({
            position: {
                top:50
            },
            title: title,
            //visible: false,
            modal: false,
            close: typeof fnClose=="function" ? 
                fnClose :
                function () { }
        });
        kw = ow.data("kendoWindow");
        kw.open();
    }
    else {
        kw = ow.data("kendoWindow");
        kw.open();
    }
    ow.show();
    kw.center();
}

function goto(url)
{
    location.href = url;
}

function json2observableArray(url, data, arrayObject, fnOk) {
    var koResult = "";
    $.ajax({
        url: url,
        cache: false,
        type: 'POST',
        data: ko.mapping.toJSON(data),
        contentType: "application/json; charset=utf-8",
        success: function (data) {
            arrayObject.removeAll();
            if (data.length > 0) {
                data.forEach(function (x) {
                    x = ko.mapping.fromJS(x);
                    arrayObject.push(x);
                });
            }
            if (typeof fnOk == "function") fnOk(data);
            if (status != undefined) status.value = "OK";
        },
        error: function (error) {
            if (status != undefined) status.value = error.responseText;
            alert("There was an error posting the data to the server: " + error.responseText);
        }
    });

    return koResult;
}

function json2observable(url, data, observableObject, fnOk, dataProperty) {
    var koResult = "";
    $.ajax({
        url: url,
        cache: false,
        type: 'POST',
        data: ko.mapping.toJSON(data),
        contentType: "application/json; charset=utf-8",
        success: function (data) {
            var dataObj = data;
            if (dataProperty != undefined && data.hasOwnProperty(dataProperty))
            {
                dataObj = data[dataProperty];
            }
            var koObj = null;
            if (observableObject == null || observableObject == undefined) {
                koObj = ko.mapping.fromJS(dataObj);
                koResult = ko.observable(koObj);
            }
            else {
                var koObj = ko.mapping.fromJS(dataObj);
                observableObject(koObj);
            }
            if (typeof fnOk == "function") fnOk(dataObj);
        },
        error: function (error) {
            alert("There was an error posting the data to the server: " + error.responseText);
        }
    });

    return koResult;
}

function assignObservableArray(obsArrayObj, data, simpleAssign, addEvent, itemModel) {
    obsArrayObj.removeAll();
    var useSimpleAssign = arguments.length > 2 ? simpleAssign : false;
    if (typeof data != "undefined" && data.length > 0) {
        if (useSimpleAssign) {
            data.forEach(function (x) {
                if (typeof addEvent == "function") addEvent(x);
                obsArrayObj.push(x);
            });
        }
        else {
            data.forEach(function (x) {
                if (itemModel == undefined) {
                    x = ko.mapping.fromJS(x);
                }
                else {
                    x = ko.mapping.fromJS(x, itemModel);
                }
                if (typeof addEvent == "function") addEvent(x);
                obsArrayObj.push(x);
            });
        }
    }
}

function push2ObservableArray(obsArrayObj, data) {
    obsArrayObj.removeAll();
    data.forEach(function (x) {
        obsArrayObj.push(x);
    });
}

function ajaxPost(url, data, fnOk, fnNok) {
    $.ajax({
        url: url,
        type: 'POST',
        data: ko.mapping.toJSON(data),
        contentType: "application/json; charset=utf-8",
        success: function (data) {
            if (typeof fnOk == "function") fnOk(data);
            koResult = "OK";
        },
        error: function (error) {
            if (typeof fnNok == "function") {
                fnNok(error);
            }
            else {
                alert("There was an error posting the data to the server: " + error.responseText);
            }
        }
    });
}

function coalesce(nullcheckobj, defaultvalue) {
    return nullcheckobj == null || nullcheckobj == undefined ?
        defaultvalue : nullcheckobj;
}

function kendoGridRezize(gridObject, newSize) {
    var gridElement = gridObject;
    var dataArea = gridElement.find(".k-grid-content");
    var newHeight = newSize;
    var diff = gridElement.innerHeight() - dataArea.innerHeight();
    gridElement.height(newHeight);
    dataArea.height(newHeight - diff);
}

function openWinFrame(frameId, frameUrl, options) {
    var bd = $("body");
    var frameId_selector = "#" + frameId;

    if (options==undefined)
    {
        options = {
            frameTitle: frameId,
            callbackClose: undefined,
            width: 600,
            height: 600
        }
    }
    if (!options.hasOwnProperty("frameId")) options.frameId = frameId;
    if (!options.hasOwnProperty("callback_close")) options.callback_close = undefined;
    if (bd.find(frameId_selector).length == 0) {
        html = "<div id='" + frameId + "' style='display:none;" +
            "width:" + (options.hasOwnProperty("width") ? options.width : 600) + "px;" +
            "height:" + (options.hasOwnProperty("height") ? options.height : 600) + "px;'>" +
            "<iframe id='ifrm_" + frameId + "' src='' scrolling='auto' frameborder='no' style='width:100%;height:100%' />" +
            "</div>";
        bd.append(html);
    }
    showDialog($(frameId_selector), options.frameTitle, options.callbackClose);
    $("#ifrm_" + frameId).prop("src", frameUrl);
}

function setDs(bindtoObject, dataIndex, ds) {
    $.each($("[bindto='"+bindtoObject+"']"), function (idx, obj) {
        var jobj = $(obj);
        if (jobj.data(dataIndex)) {
            jobj.data(dataIndex).setDataSource(ds);
        }
    });
}

function BSTabShow(tab) {
    $('.tab-pane a[href="#' + tab + '"]').tab('show');
};

function openModule(module_type, module_id, module_parm, removeIt) {
    var menuObj = getMenuObj(module_type, module_id);
    var appContent = $('.appContent[module_id="' + module_id + '"]');
    var appContentOther = $('.appContent:not([module_id="' + module_id + '"])');
    if (menuObj != undefined) {
        if (menuObj.hasOwnProperty("main_controller")) linkToOpen = menuObj.main_controller;
        if (linkToOpen == "" && module_id != "Home") return;
        if (module_parm != undefined) {
            linkToOpen += linkToOpen.indexOf("?") > 0 ? "&" : "?";
            linkToOpen += "parm=" + module_parm;
        }
        if (module_id != "Home") {
            var exist = false;
            if(removeIt===true && appContent.length>0)
            {
                appContent.remove();
            }
       
            if (appContent.length==0 || removeIt===true)
            {
                $("#appContainer").append('<div class="appContent" module_id="' + module_id + '"><iframe width="100%" height="100%" frameborder="yes" /></div>');
                appContent = $('.appContent[module_id="' + module_id + '"]');
                appContent.find("iframe").attr("src", linkToOpen);
            }
            else
            {
                exist = true;
            }
            
            if (menuObj.hasOwnProperty("reload_when_open") || exist===false)
            {
                if (menuObj.reload_when_open === true) {
                    appContent.find("iframe").attr("src", linkToOpen);
                }
            }
        }
    }

    if (menuObj != undefined || module_id == "Home") {
        if (module_id != "Home") {
            var winHeight = $(window).height() - 60;
            //appContent.css("border","solid 1px #000");
            appContent.width("100%").height($(window).height()-120);
            appContent.find("iframe").width("100%").height("100%");
        }
        appContentOther.stop().slideUp("slow");
        appContent.stop().slideDown("slow");
    }
}

function linq2array(sources,fnValidate, fnGet)
{
    if (fnValidate == undefined)
    {
        fnValidate = function () {
            return true;
        }
    }

    if (fnGet == undefined)
    {
        fnGet = function(obj)
        {
            return obj;
        }
    }

    var as = new Array();
    sources.ForEach(function (obj, idx) {
        if (fnValidate(obj)) as.push(fnGet(obj));
    });
    return as;
}

function progressingChart(objs)
{
    $.each(objs, function (i, o) {
        var obj = $(o);
        if (obj.children(".chart_progress").length == 0) {
            $("<div style='width:100%;text-align:center' class='chart_progress'>"
                //+ "<img src='~/Content/Images/loader1.gif' />"
                + "<br />Please wait while loading ..."
                + "</div>").show().appendTo(obj);
        }

        if (obj.children(".chart_canvas").length == 0) {
            $("<div style='width:100%;height=100%' class='chart_canvas'>"
                + "</div>").hide().appendTo(obj);
        }
    });

    $.each(objs, function (i, o) {
        var obj = $(o);
        progress = obj.find(".chart_progress");
        canvas = obj.find(".chart_canvas");
        progress.show();
        canvas.hide();
    });
}

function runScripts(scripts) {
    scripts.forEach(function (s, idx) {
        if (typeof s == "function") {
            s();
        }
    });
}

function drawCanvasChart(objs, data, options) {
   $.each(objs, function (i, o) {
        var obj = $(o);
        progress = obj.find(".chart_progress");
        canvas = obj.find(".chart_canvas");
        if(progress!=undefined)progress.hide();
        if (canvas != undefined) canvas.show();
        canvas.width(obj.width()).height(obj.height());
        if (data == undefined)
        {
            canvas.html("<center>Unable to create chart<br/>Returned data is not valid to draw chart<center>");
            return;
        }

        if (options == undefined) {
            canvas.html("<center>Unable to create chart<br/>Chart definition is not valid to draw chart<center>");
            return;
        }

        if (data.hasOwnProperty("Result") && data["Result"] != "OK") {
            var msg = data.hasOwnProperty("Message") ? data["Message"] + "<br/>"+ data["Trace"] : "Error";
            canvas.html("<center>Unable to create chart<br/>" + msg + "<center>");
        }
        else {
            if (typeof options == "function") {
                canvas.kendoChart(options(data));
            }
            else
            {
                canvas.kendoChart(options);
            }
        }
    });
}

function writeToChart(objs, htmlText) {
    $.each(objs, function (i, o) {
        var obj = $(o);
        progress = obj.find(".chart_progress");
        canvas = obj.find(".chart_canvas");
        progress.hide();
        canvas.show();
        canvas.html(htmlText);
    });
}

function makeChart(objs, ajaxUri, ajaxParms, chartOptions, callbackAfter) {
    //--- build obj child
    progressingChart(objs);

    var progress, canvas;
    if (ajaxUri != undefined) {
        if (ajaxParms == undefined) ajaxParms = "";
        ajaxPost(ajaxUri, ajaxParms,
            function (data) {
                $.each(objs, function (i, o) {
                    /*
                    var obj = $(o);
                    progress = obj.find(".chart_progress");
                    canvas = obj.find(".chart_canvas");
                    progress.hide();
                    canvas.show();
                    canvas.width(obj.width()).height(obj.height());
                    //canvas.html("<center>Chart will be displayed here<center>");
                    if (data.hasOwnProperty("Result") && data["Result"] != "OK") {
                        var msg = data.hasOwnProperty("Message") ? data["Message"] + data["Trace"]: "Error";
                        canvas.html("<center>Unable to create chart<br/>" + msg + "<center>");
                    }
                    else {
                        canvas.kendoChart(chartOptions(data));
                    }
                    */
                    drawCanvasChart(objs, data, chartOptions);
                });
            },
            function (err) {
                writeToChart(objs,"<center>Unable to create chart<br/>" + err.responseText + "<center>");
            })
    }
    else {
        $.each(objs, function (i, o) {
            var obj = $(o);
            progress = obj.find(".chart_progress");
            canvas = obj.find(".chart_canvas");
            progress.hide();
            canvas.html("<center>Unable to create chart<br/>Invalid controller<center>");
        });
    }
}

var downloadURL = function downloadURL(url) {
    var hiddenIFrameID = 'hiddenDownloader',
        iframe = document.getElementById(hiddenIFrameID);
    if (iframe === null) {
        iframe = document.createElement('iframe');
        iframe.id = hiddenIFrameID;
        iframe.style.display = 'none';
        document.body.appendChild(iframe);
    }
    iframe.src = url;
};

function showApp(id) {
    id = id.replace(" ", "_");
    var appContainer = $("#ecis_app_container");
    var panelApp = appContainer.children("#ecis_app_" + id)[0];
    if (panelApp == undefined) {
        panelApp = $("<div id='ecis_app_" + id + "' class='ecis_panel'></div>").appendTo(appContainer);
    }
    else {
        panelApp = $(panelApp);
    }

    appContainer.children().each(function (idx, obj) {
        obj = $(obj);
        if (obj.is("#ecis_app_" + id)) {
            obj.show();
        }
        else {
            obj.hide();
        }
    });
}

function loadApp(id, url, parm) {
    id = id.replace(" ", "_");
    id = "panel";
    var appContainer = $("#ecis_app_container");
    var panelApp = appContainer.children("#ecis_app_" + id)[0];
    if (panelApp == undefined) {
        panelApp = $("<div id='ecis_app_" + id + "' class='ecis_panel'></div>").appendTo(appContainer);
    }
    else
    {
        panelApp=$(panelApp);
    }

    appContainer.children().each(function (idx, obj) {
        obj=$(obj);
        if(obj.is("#ecis_app_" + id))
        {
            obj.show();
        }
        else
        {
            obj.hide();
        }
    });
    panelApp.html(
        "<center>" +
        //"<img src=\"~/Content/Images/loader1.gif\" />" +
        "<br />Please wait while loading</center>");
    $("#AppLeft").html("");
    ko.cleanNode($("#AppLeft")[0]);

    ajaxPost(url, parm, function (data) {
        if (data.Result == 'OK') {
            if (data.hasOwnProperty("Data")) {
                var panelData = data.Data;
                if (panelData.hasOwnProperty("LeftHtml")) {
                    $("#AppLeft").html(panelData.LeftHtml);
                }
                if (panelData.hasOwnProperty("Html")) {
                    panelApp.html(panelData.Html);
                }
                //ko.cleanNode($(document));
                //ko.applyBindings(model,$(""));
            }
            else {
                panelApp.html("<center>Module has been loaded</center>");
            }
        }
        else {
            var msg = data.hasOwnProperty("Message") ? data.Message + data.Trace : "Error loading, please contact administrator";
            panelApp.html("<center>Unable to load module<br/><br/>" + msg + "</center>");
        }
    },
    function (err) {
        panelApp.html("<center>Unable to load module<br/><br/>" + err.responseText + "</center>");
    });
}

function componentToHex(c) {
    var hex = c.toString(16);
    return hex.length == 1 ? "0" + hex : hex;
}

function rgbToHex(r, g, b) {
    return "#" + componentToHex(r) + componentToHex(g) + componentToHex(b);
}
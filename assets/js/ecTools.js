function ks(o) {
    return kendo.stringify(o);
}

function kf(f,v) {
    return kendo.format(f,v);
}

function gp(obj, propName, def) {
    var r = obj.hasOwnProperty(propName) ? obj[propName] : def;
    return r;
}

function n2s(obj) {
    var fmt = "{0:N0}";
    return kendo.format(fmt,obj);
}

function fd(format, value, def) {
    if (def == undefined) def = "-";
    var ret = "";
    if (value == 0 || value == "") {
        ret = def;
    } else {
        ret = kendo.format(format, value);
    }
    return ret;
}

function tooggleCheckbox(obj, selectorTxt) {
    var cbxs = $(selectorTxt);
    var checked = obj.prop("checked");
    cbxs.prop("checked", checked);
}

function objectScript(fn, group) {
    var o = this;
    o.fn = fn;
    o.group = group;
    return o;
}

function addObjScript(obj, scr, group) {
    if (obj.data("ecscript") == undefined) obj.data("ecscript", []);
    var ss = obj.data("ecscript");
    if (typeof scr == "function") ss.push(new objectScript(scr, group == undefined ? "default" : group));
}

function runObjScript(obj, g) {
    if (obj.data("ecscript") != undefined) {
        var ss = obj.data("ecscript");
        if (g == undefined) g = "default";
        ss.forEach(function (obj) {
            if (obj.group == g) obj.fn();
        });
    }
}

function showErr(e, fn) {
    //alert(e.responseText);
    var errMsg = "";
    model.Processing(false);
    if (typeof e == "string") errMsg = e;
    else if (e.hasOwnProperty("responseText")) errMsg = e.responseText;
    else if (e.hasOwnProperty("Message") && e.hasOwnProperty("Trace")) errMsg = e.Message + "\n" + e.Trace;
    else if (e.hasOwnProperty("Message")) e = e.Message;
    alert(errMsg);
    if (typeof fn == "function") fn(e);
}

//--- Kendo Related funciton
function makeField(label,inputDataBind,labelClass,fieldClass,inputClass) {
    var str = "<label class=\"" + labelClass + "\">" + label + "</label>\n";
    str += "<div class=\""+fieldClass+"\"><input class=\""+inputClass+"\" data-bind=\""+inputDataBind+"\"></div>";
    document.write(str);
}

function normalizeTree(objs, el, parentPreText, fnTr) {
    var ret = [];
    objs.forEach(function (obj) {
        var it = fnTr(obj, parentPreText);
        ret.push(it);
        if (obj.hasOwnProperty(el)) {
            var newParentPreText = parentPreText + it.shortTitle + " \\ ";
            //alert(it.title + " ==> " + newParentPreText);
            var childs = normalizeTree(obj[el], el, newParentPreText, fnTr);
            childs.forEach(function (c) {
                ret.push(c);
            });
        }
    });
    return ret;
}

function input2datePicker(objects) {
    $.each(objects, function (idx, obj) {
        var jobj = $(obj);
        var dateval = jobj.val();
        if (jobj.data("kendoDatePicker") == undefined) {
            var fmt = jobj.attr("format") == undefined ? jsonDateFormat : jobj.attr("format");
            var depth = jobj.attr("depth") == undefined ? "month" : jobj.attr("depth");
            var start = jobj.attr("start") == undefined ? "month" : jobj.attr("start");
            var min = jobj.attr("min") == undefined ? new Date(1900,1,1) : jsonDate(jobj.attr("min"));
            var max = jobj.attr("max") == undefined ? new Date(3000, 12, 31) : jsonDate(jobj.attr("max"));
            jobj.kendoDatePicker({
                format: fmt, start: start,
                min: min, max:max,
                depth: depth, parseFormats: ["dd-MMM-yyyy"]
            });
        }
        jobj.data("kendoDatePicker").value(jsonDateStr(dateval));
    });
}
//--- End of Kendo

//--- Date related function
function jsonDate(strDt) {
    if (strDt == undefined) return "";
    var dt = str2date(strDt);
    if (dt.getFullYear() <= 1970 || dt.getFullYear() == 1) dt = "";
    return dt;
}

function jsonDateStr(dtSource, format) {
    var dt = str2date(dtSource);
    if (dt == null || dt == undefined || dt == "") return "";
    if (dt.getFullYear() <= 1970 || dt.getFullYear() == 1) return "";
    var ret = kendo.toString(dt, format == undefined ? jsonDateFormat : format);
    if (ret.indexOf("NaN") >= 0) return "";
    return ret;
}

//alert(jsonDateStr("25-Mar-2014"));

function str2date(dtSource) {
    if (dtSource == null) return "";
    dtSource = dtSource.toString();
    var dt = dtSource;
    if (dtSource.substr(0, 6) == "/Date(") {
        var dtParse = Date.parse(dtSource);
        if (isNaN(dtParse)) {
            var intMs = parseInt(dtSource.substr(6));
            dt = new Date(intMs);
        }
        else {
            dt = new Date(dtParse);
        }
        //alert(dt);
        dt = new Date(dt.getTime() + dt.getTimezoneOffset() * 60000);
    }
    else if (dtSource.length == 5 && dtSource.substr(2, 1) == ":") {
        var times = dtSource.split(":");
        dt = new Date();
        dt = new Date(dt.getFullYear(), dt.getMonth(), dt.getDate(), times[0], times[1]);
    }
    else {
        dt = new Date(dtSource);
        if (!isDate(dt)) dt = jsonDateFromStr(dtSource);
        //dt = new Date(dt.getTime() + dt.getTimezoneOffset() * 60000);
    }
    return dt;
}

function isDate(dt) {
    return (dt instanceof Date && !isNaN(dt.valueOf()));
}

function date2str(dt, format) {
    if (dt instanceof Date === false) return "";
    if (dt.getFullYear() <= 1970 || dt.getFullYear() == 1) return "";
    return kendo.toString(dt, format == undefined ? jsonDateFormat : format);
}

//ie: 01-Jan-2013
function jsonDateFromStr(strDt) {
    var yr = parseInt(strDt.substr(7, 4));
    var mth = strDt.substr(3, 3);
    var day = parseInt(strDt.substr(0, 2));
    switch (mth) {
        case "Jan": mth = 0; break;
        case "Feb": mth = 1; break;
        case "Mar": mth = 2; break;
        case "Apr": mth = 3; break;
        case "May": mth = 4; break;
        case "Jun": mth = 5; break;
        case "Jul": mth = 6; break;
        case "Aug": mth = 7; break;
        case "Sep": mth = 8; break;
        case "Oct": mth = 9; break;
        case "Nov": mth = 10; break;
        case "Dec": mth = 11; break;
    }
    var dt = new Date(yr, mth, day);
    if (dt.getFullYear() == 1900) dt = "";
    return dt;
}

function toUTC(dt) {
    if (dt == undefined || dt == "") return "";
    dt = new Date(dt.getTime() + dt.getTimezoneOffset() * 60000);
    return dt;
}
//--- end of date related function

function obsArray2Js(obs) {
    var ret = [];
    obs().forEach(function (obj) {
        ret.push(ko.mapping.toJS(obj));
    });
    return ret;
}

function getObjectProperties(obj) {
    var keys = [];
    for (var key in obj) {
        keys.push(key);
    }
    return keys;
}

function jsonObjsConvDate(as, dateOrStr) {
    as.forEach(function (e) {
        e = jsonObjConvDate(e, dateOrStr);
    });
    return as;
}

function jsonObjConvDate(e, dateOrStr) {
    if (dateOrStr == undefined) dateOrStr = "date";
    var keys = getObjectProperties(e);
    keys.forEach(function (k) {
        if (typeof e[k] == "string" && e[k] != null && e[k] != undefined) {
            if (e[k].indexOf("/Date") >= 0) {
                var dt = dateOrStr=="str" ? jsonDateStr(e[k]) : jsonDate(e[k]);
                e[k] = dt;
            }
        }
        else if (typeof e[k] == "object") {
            e[k] = jsonObjConvDate(e[k]);
        }
    });
    return e;
}

function ajaxPost(url, data, fnOk, fnNok) {
    $.ajax({
        url: url,
        type: 'POST',
        data: ko.mapping.toJSON(data),
        contentType: "application/json; charset=utf-8",
        success: function (data) {
            data = jsonObjConvDate(data);
            if (typeof fnOk == "function") fnOk(data);
            koResult = "OK";
        },
        error: function (error) {
            model.Processing(false);
            if (typeof fnNok == "function") {
                fnNok(error);
            }
            else {
                alert("There was an error posting the data to the server: " + error.responseText);
            }
        }
    });
}


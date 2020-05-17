function ecFormField(obj) {
    var othis = this;
    othis.colspan = obj.hasOwnProperty("colspan") ? obj.colspan : 12;
    othis.colclass = "col-sm-" + this.colspan.toString();
    othis.fields = obj.hasOwnProperty("fields") ? obj.fields : [];
    othis.control = (obj.hasOwnProperty("control") ? obj.control : "").toString().toLowerCase();
    othis.modeField = gp(obj, "modeField", "currentMode()");
    
    othis.render = function (context, parentDiv) {
        if (parentDiv.attr("column") != undefined) {
            othis.colspan = Math.floor(12 / parentDiv.attr("column"));
            othis.colclass = "col-sm-" + othis.colspan.toString();
        }

        var f = $("<div class=\"ec_df_fieldcell " + othis.colclass + " " +
            "\" id=\"" + obj.fieldId + "\"></div>");
        switch (othis.control) {
            case "tabcontainer":
                othis.renderTabContainer(obj, context, f);
                break;

            case "tab":
                othis.renderTab(obj, context, parentDiv);
                break;

            case "group":
                othis.renderGroup(obj, context, f);
                break;

            case "treeview":
                othis.renderTreeView(obj, context, f);
                othis.fields = [];
                break;

            default:
                othis.renderElement(obj,context,f);
                break;
        }

        var fieldContainer = undefined;
        if (othis.fields.length > 0) {
            if (othis.control != "tabcontainer" && othis.control!="tab") {
                fieldContainer = $("<div class=\"row\"></div>");
                if (gp(obj, "column", 1) > 1) fieldContainer.attr("column", gp(obj, "column", 1));
                othis.fields.forEach(function (fld) {
                    fld.modeField = obj.modeField;
                    new ecFormField(fld).render(context, fieldContainer);
                });
                fieldContainer.appendTo(f);
            }
            else if (othis.control == "tab") {
                fieldContainer = parentDiv.find("div#tab-pane-" + obj.fieldId + ":first");
                if (gp(obj, "column", 1) > 1) fieldContainer.attr("column", gp(obj, "column", 1));
                if (fieldContainer != undefined) {
                    othis.fields.forEach(function (fld) {
                        new ecFormField(fld).render(context, fieldContainer);
                    });
                }
            }
            else {
                othis.fields.forEach(function (fld) {
                    fld.modeField = 
                    new ecFormField(fld).render(context, f);
                });
                f.find("ul.nav-tabs").find("li:first").addClass("active");
                f.find("div.tab-content").find("div.tab-pane:first").addClass("active");
            }
        }

        if (othis.control != "tab")f.appendTo(parentDiv);
    }

    othis.renderTabContainer = function (obj, context, parentDiv) {
        var vi = [gp(obj, "visible", true), gp(obj, "visibleNew", true), gp(obj, "visibleEdit", true)];
        var dbVisible = "visible:true";
        if (vi[0] == false) dbVisible += " && true";
        //if (vi[1] == false) dbVisible += " && "+modeField+"=='Edit'";
        //if (vi[2] == false) dbVisible += " && "+modeField+"=='New'";

        $("<ul class=\"nav nav-tabs\"></ul>").appendTo(parentDiv);
        $("<div class=\"tab-content\" id=\"tab-content-"+obj.fieldId+"\"></div>").appendTo(parentDiv);
    }

    othis.renderTab = function (obj, context, parentDiv) {
        var vi = [gp(obj, "visible", true), gp(obj, "visibleNew", true), gp(obj, "visibleEdit", true)];
        var dbVisible = "visible:true";
        if (vi[0] == false) dbVisible += " && true";
        //if (vi[1] == false) dbVisible += " && "+modeField+"=='Edit'";
        //if (vi[2] == false) dbVisible += " && "+modeField+"=='New'";

        var tabheader = parentDiv.find("ul.nav-tabs:first");
        if (tabheader != undefined) {
            tabheader.append($("<li><a href=\"#tab-pane-"+obj.fieldId+"\" data-toggle=\"tab\">"+obj.title+"</a></li>"));
        }
        var tabcontent = parentDiv.find("div.tab-content:first");
        if (tabcontent != undefined) {
            tabcontent.append($("<div class=\"tab-pane\" id=\"tab-pane-"+obj.fieldId+"\" style=\"margin-bottom:5px;\"></div>"));
        }
    }

    othis.renderTreeView = function (obj, context, parentDiv) {
        var vi = [gp(obj, "visible", true), gp(obj, "visibleNew", true), gp(obj, "visibleEdit", true)];
        var dbVisible = "visible:true";
        if (vi[0] == false) dbVisible += " && true";
        //if (vi[1] == false) dbVisible += " && "+modeField+"=='Edit'";
        //if (vi[2] == false) dbVisible += " && "+modeField+"=='New'";
        if (gp(obj, "title", "") != "")
            $("<div><h5>" + obj.title + "</h5></div>").appendTo(parentDiv);
        $("<div id=\"TreeView-" + obj.fieldId + "\" class=\"ec-df-treeview\" " +
            "datasource=\""+ gp(obj, "dataSource", "") + "\" " +
            //"style=\"border:solid 1px #ddd;height:400px\" " +
            "></div>")
            .appendTo(parentDiv);
    }

    othis.renderGroup = function (obj, context, parentDiv) {
        var vi = [gp(obj, "visible", true), gp(obj, "visibleNew", true), gp(obj, "visibleEdit", true)];
        var modeField = obj.modeField;
        var dbVisible = "visible:true";
        if (vi[0] == false) dbVisible += " && true";
        if (vi[1] == false) dbVisible += " && "+modeField+"=='Edit'";
        if (vi[2] == false) dbVisible += " && "+modeField+"=='New'";

        var objForm = ($("<div class=\"ec_df_form_group row\" data-bind=\"" + dbVisible + "\"></div>"))
            .appendTo(parentDiv);
        var header = $("<div class=\"col-md-12\"><h3>" + gp(obj,"title","") + "</h3></div>");
        header.appendTo(objForm);
    }

    othis.renderElement = function (obj, context, parentDiv) {
        var dataSources = context.dataPoints;
        var controlType = gp(obj, "control", "text");
        var vi = [gp(obj, "visible", true), gp(obj, "visibleNew", true), gp(obj, "visibleEdit", true)];
        var ro = [gp(obj, "readOnly", false), gp(obj, "readOnlyNew", false), gp(obj, "readOnlyEdit", false)];

        //alert(ks(context));

        var textInfo = "";
        var input = "";
        var dbVisible = "";
        var dbRO = "";
        var modeField = obj.modeField;

        var ds = undefined;
        if (obj.hasOwnProperty("control")) controlType = obj.control;
        if (obj.hasOwnProperty("dataSource") && dataSources != undefined) {
            ds = gp(dataSources, obj.dataSource, undefined);
        }

        if (vi[0] == true) {
            dbVisible = ""+modeField+"==''";
            if (dbVisible != "") dbVisible = "visible:" + dbVisible;
            if (ro[0] == true) dbVisible += " || "+modeField+"!=''";
            if (ro[1] == true) dbVisible += " || "+modeField+"=='New'";
            if (ro[2] == true) dbVisible += " || "+modeField+"=='Edit'";
            textInfo = "<span " +
                "data-bind=\"text:" + obj.field + "," + dbVisible +
                "\"" +
                 "/>";
         
            if (ro[0] != true) {
                dbRO = "visible:true";
                if (ro[0] == true) dbRO += " && false";
                if (ro[1] == true) dbRO += " && "+modeField+"=='Edit'";
                if (ro[2] == true) dbRO += " && "+modeField+"=='New'";

                var dbValue = obj.hasOwnProperty("field") ? "value:" + obj.field : "";
                var style = gp(obj,"style","");
                if (obj.hasOwnProperty("width")) style += "width:" + obj.width;
                if (style != "") style = "style=\""+style+"\" ";

                switch (controlType) {
                    case "select":
                        var os = "";
                        if (ds!=undefined) {
                            var dsData = ds.run();
                            dsData.forEach(function (i, v) {
                                if (typeof i == "string" || isNaN(i) == false) {
                                    os += "<option value='" + i + "'>" + i + "</option>\n";
                                }
                            });
                        }
                        input = "<select type='text' placeholder='Please enter " + obj.title + "' " +
                            style +
                            "data-bind=\"" + dbValue + "," + dbRO +
                            "\"" + //-- end of data-bind
                            ">" + os + "</select>";
                        break;

                    case "date":
                        input = "<input type='text' control='datetime' class='ec_df_entry ec_df_date'  " +
                        "placeholder='Please enter " + obj.title + "' " +
                        style +
                        "data-bind=\"" + dbValue + "," + dbRO +
                        "\"" +
                        ">";
                        break;

                    default:
                        input = "<input type='text' class='ec_df_input' " +
                        style +
                        "placeholder='Please enter " + obj.title + "' " +
                        "data-bind=\"" + dbValue + "," + dbRO +
                        "\"" +
                        ">";
                }
            }

            dbVisible = "visible:true";
            if (vi[0] == false) dbVisible += " && true";
            if (vi[1] == false) dbVisible += " && "+modeField+"=='Edit'";
            if (vi[2] == false) dbVisible += " && "+modeField+"=='New'";
        }
        else {
            dbVisible = "visible:false";
        }
        //alert(dbVisible);

        if (obj.hasOwnProperty("template")) {
            var t = $("#" + obj.template);
            if (t != undefined) {
                if (objForm != undefined) {
                    var d = $("<div class=\"ec_df_fieldcell col-sm-12\" data-bind=\"" + dbVisible + "\"></div>").appendTo(parentDiv);
                    d.html(t.html());
                    return d;
                };
            }
            else {
                return $("");
            }
        }

        f = $("<label class='col-md-2 align_right'>" + obj.title + "</label>" +
                "<div class='col-md-10'>" +
                textInfo +
                input +
                "</div>");

        f.appendTo(parentDiv);
    }
}
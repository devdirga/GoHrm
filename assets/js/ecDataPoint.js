function ecDataPoint(args) {
    var othis = this;
    this.id = undefined;
    this.uri = undefined;
    this.uriData = undefined;
    this.data = undefined;
    this.preRun = undefined;
    this.postRun = undefined;

    switch (typeof args) {
        case "string":
            this.uri = args;
            break;

        case "object":
            if (args.hasOwnProperty("id")) this.id = args.id;
            if (args.hasOwnProperty("uri")) this.uri = args.uri;
            if (args.hasOwnProperty("uriData")) this.uriData = args.uriData;
            if (args.hasOwnProperty("data")) this.data = args.data;
            if (args.hasOwnProperty("preRun")) this.preRun = args.preRun;
            if (args.hasOwnProperty("postRun")) this.postRun = args.postRun;
            break;

        case "array":
            this.data = args;
    }

    this.run = function (m, runOK) {
        if (othis.uri == undefined && othis.data == undefined) throw "Data Point URI (" + othis.uri + ") is invalid";
        
        if (othis.data != undefined) return othis.data;
        var postData = {};
        if (othis.uriData != undefined) {
            switch (typeof othis.uriData) {
                case "function":
                    postData = othis.uriData(m);
                    break;

                default:
                    postData = othis.uriData;
                    break;
            }
        }
        return ajaxPost(othis.uri, postData, function (data) {
            if (data.hasOwnProperty("Result")) {
                if (data.Result == "OK") {
                    if (typeof othis.postRun == "function") {
                        othis.postRun(data);
                    }
                    if (typeof runOK == 'function') {
                        var ok = runOK(data);
                        //alert(othis.uri + " r:" + kendo.stringify(ok));
                        return ok;
                    }
                    else {
                        //alert(ks(data));
                        return data;
                    }
                }
                else {
                    throw data.Message;
                }
            }
            else {
                throw "Uknown error";
            }
        },
        function showerr(e) {
            throw e;
        });
    }
}
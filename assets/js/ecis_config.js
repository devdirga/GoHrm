//--- ECIS General Configuration
var ecisBaseUrl = window.location.protocol + "//"
        + window.location.host + "/"
        + window.location.pathname.split('/')[1]
        + "/";

var ecisHost = window.location.protocol + "//"
        + window.location.host + "/";

var ecisBaseUrl_noLastTrail = window.location.protocol + "//"
        + window.location.host + "/"
        + window.location.pathname.split('/')[1];

var ecisHost_noLastTrail = window.location.protocol + "//"
        + window.location.host;

var jsonDateFormat = "dd-MMM-yyyy";
//var cultureInfo = "en-SG";

var ecisColors = ["#317DB5", "#73B576", "#E5D159", "#F7A55C", "#D51D26",
"#128572", "#8B7FC0", "#747474", "#D85F21", "#7CC8C0"];
 
var ecisColors2 = ["#9DA7E0", "#19E083", "#191CE0", "#E0195F", "#E3FC3F",
"#2DFAEC", "#FF7E38", "#34D100", "#B0B0B0", "#F51DE6"];

function getEcisColor(iColorIndex) {
    return getArrayElement(ecisColors, iColorIndex);
}

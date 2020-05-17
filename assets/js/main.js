var scripts = document.getElementsByTagName("script");
var mainjs = scripts[scripts.length - 1].src;
var jspaths = mainjs.split("/");
jspaths.pop();
var jspath = jspaths.join("/") + "/";

document.writeln("<script src=\"" + jspath + "ecTools.js\"></script>");
document.writeln("<script src=\"" + jspath + "ecAppEngine.js\"></script>");
document.writeln("<script src=\"" + jspath + "ecDataPoint.js\"></script>");

document.writeln("<link rel='stylesheet' href=\"" + jspath + "../css/ec_df.css\"></link>");
document.writeln("<script src=\"" + jspath + "ecDataFormBuilder.js\"></script>");
document.writeln("<script src=\"" + jspath + "ecDataForm.js\"></script>");
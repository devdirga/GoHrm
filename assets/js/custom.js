function formatNumber(numb, format) {
    var ret = 0;
    // if(parseFloat(numb) >= Math.pow(10, 9)) {
    //     ret = kendo.toString(parseFloat(numb)/(Math.pow(10, 9)), format) + ' B'; 
    // }
    // else if(parseFloat(numb) >= Math.pow(10, 6)) {
    //     ret = kendo.toString(parseFloat(numb)/(Math.pow(10, 6)), format) + ' M'; 
    // }
    // else if(parseFloat(numb) >= Math.pow(10, 3)) {
    //     ret = kendo.toString(parseFloat(numb)/(Math.pow(10, 3)), format) + ' K'; 
    // }
    // else {
    //     ret = kendo.toString(parseFloat(numb), format);    
    // }
     ret = kendo.toString(parseFloat(numb)/(Math.pow(10, 6)), format); 

    return ret;
}

function NormalizeNumber(numb, format) {
    var ret = 0;
    if(parseFloat(numb) >= Math.pow(10, 9)) {
        ret = kendo.toString(parseFloat(numb)/(Math.pow(10, 9)), format) + ' B'; 
    }
    else if(parseFloat(numb) >= Math.pow(10, 6)) {
        ret = kendo.toString(parseFloat(numb)/(Math.pow(10, 6)), format) + ' M'; 
    }
    else if(parseFloat(numb) >= Math.pow(10, 3)) {
        ret = kendo.toString(parseFloat(numb)/(Math.pow(10, 3)), format) + ' K'; 
    }
    else {
        ret = kendo.toString(parseFloat(numb), format);    
    }

    return ret;
}
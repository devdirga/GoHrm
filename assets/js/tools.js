var Tools = {};

Tools.Util = {
    redefine: function (what, defaultValue) {
        if (typeof what === 'undefined') {
            return (typeof defaultValue === 'undefined' ? '' : defaultValue);
        }

        return what;
    }
}

Tools.Date = {
    names: function (what, locale, startFrom) {
        what = Tools.Util.redefine(what, 'days');
        locale = Tools.Util.redefine(locale, 'en');

        var data = {
            en: {
                days: ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday']
            }
        };

        try {
            var selectedWhat = data[locale][what];
            startFrom = Tools.Util.redefine(startFrom, selectedWhat[0]);
            var shiftLength = selectedWhat.indexOf(startFrom);

            if (shiftLength > 0) {
                return selectedWhat.concat(selectedWhat.splice(0, shiftLength));
            }

            return selectedWhat;
        } catch (err) {
            return [];
        }
    },
    diffMonth: function (date1, date2) {
        var months;

        months = (date2.getFullYear() - date1.getFullYear()) * 12;
        months -= date1.getMonth() + 1;
        months += date2.getMonth();

        return (months <= 0 ? 0 : months);
    },
    fromCSharpDateTime: function(cSharpDateTime) {
        return new Date(parseInt((/-?\d+/).exec(cSharpDateTime)[0], 10));
    }
};

Tools.String = {
    reverse: function (string) {
        return string.split('').reverse().join('');
    },
    leftPadding: function (string, leftPadding, width) {
        string = String(string);
        string = this.reverse(string);
        width = width - string.length;

        for (i = 0; i < width; i++)
            string += leftPadding;

        return this.reverse(string);
    },
    capitalize: function (string) {
        return string && string[0].toUpperCase() + string.slice(1);
    }
};
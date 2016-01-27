viewModel.app = {}; var app = viewModel.app;

app.applyNavigationActive = function () {
	var currentURL = document.URL.split("/").slice(3).join("/");
	var $a = $("a[href='/" + currentURL + "']");

	$a.closest(".navbar-nav").children().removeClass("active");
	if ($a.closest("li").parent().hasClass("navbar-nav")) {
		$a.closest("li").addClass("active");
	} else {
		$a.closest("li").closest("li").addClass("active");
	}
};
app.prepare = function () {
	app.applyNavigationActive();
};
app.ajaxPost = function (url, data, callbackSuccess, callbackError, otherConfig) {
    var startReq = moment();
    var callbackScheduler = function (callback) {
        app.isLoading(false);
        callback();
        
        // var finishReq = moment();
        // var responseTime = finishReq.diff(startReq, "second");
        // if (responseTime > 1) {
        //     app.isLoading(false);
        //     callback();
        // } else {
        //     setTimeout(function () {
        //         app.isLoading(false);
        //         callback();
        //     }, 1000);
        // }
    };

    var config = {
        url: url,
        type: 'post',
        dataType: 'json',
        data: data,
        success: function (a) {
            callbackScheduler(function () {
                callbackSuccess(a);
            });
        },
        error: function (a, b, c) {
            callbackScheduler(function () {
                if (callbackError !== undefined) {
                    callbackError(a, b, c);
                }
            });
        }
    };

    if (data instanceof FormData) {
        config.async = false;
        config.cache = false;
        config.contentType = false;
        config.processData = false;
    }

    if (otherConfig != undefined) {
        config = $.extend(true, config, otherConfig);
    }

    app.isLoading(true);
    $.ajax(config);
};
app.isFine = function (res) {
    if (!res.success) {
        sweetAlert("Oops...", res.message, "error");
        return false;
    }

    return true;
};
app.isFormValid = function (selector) {
    $(selector).kendoValidator();
    var $validator = $(selector).data("kendoValidator");
    return ($validator.validate());
};
app.isLoading = ko.observable(false);
app.fixKendoMultiSelect = function () {
    $("*.k-multiselect").change(function (e) {
        var existingValue = $(this).attr("existingValue");
        if (existingValue !== undefined) {
            var d = $(this).find("select[multiple='multiple']").data("kendoMultiSelect");
            var newValue = existingValue.split(",");
            var existingDataSource = d.dataSource._data;
            newValue.push(d.value()[0]);
            d.setDataSource(existingDataSource);
        }
    });

    $("*.k-multiselect").keydown(function (e) {
        var d = $(this).find("select[multiple='multiple']").data("kendoMultiSelect");
        if (e.keyCode == 13) {
            var existingValue = $(this).attr("existingValue");
            var newValue = existingValue.split(",");
            var existingDataSource = d.dataSource._data;
            newValue.push(d.value()[0]);
            d.setDataSource(existingDataSource);
            d.value(newValue);
        } else {
            $(this).attr("existingValue",d.value())
        }
    });
};

$(function () {
	app.prepare();
    app.fixKendoMultiSelect();
});
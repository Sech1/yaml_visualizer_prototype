function loadGraph() {
    $.ajax({
        dataType: "json",
        type: "GET",
        url: "/api",
        success: function (data) {
            $("#yaml-graph-container").html(atob(data));
        },
        error: function (data) {
            alert(data);
        },
    });
}
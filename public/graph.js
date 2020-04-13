let data = null;

function loadGraph() {
    $.ajax({
        dataType: "json",
        type: "GET",
        url: "/api",
        success: function (json) {
            data = JSON.parse(atob(json));
            console.log(data);
            load_jsMind(data)
            $("#json-text-container").val(JSON.stringify(data));
            buildLinks();
        },
        error: function (error) {
            alert(error);
        },
    });
}

function load_jsMind(mind) {
    var options = {
        container: 'jsmind_container',
        editable: false,
        theme: 'orange'
    };
    var jm = new jsMind(options);
    // show it
    jm.show(mind);
    // jm.set_readonly(true);
    // var mind_data = jm.get_data();
    // alert(mind_data);
    //jm.add_node("sub2","sub23", "new node", {"background-color":"red"});
    //jm.set_node_color('sub21', 'green', '#ccc');
}

function buildLinks() {
    $.ajax({
        dataType: "json",
        type: "GET",
        url: "/yaml_links",
        success: function (json) {
            data = JSON.parse(atob(json));
            console.log(data);

            $("jmnode").each(function () {
                console.log($(this).attr("nodeid"));

                var link = data[$(this).attr("nodeid")];
                const originUrl = window.location.origin;
                $($(this)).on("click", function () {
                    window.open(`${originUrl}${link}`, '_blank');
                });

            });
        },
        error: function (error) {
            alert(error);
        },
    });
}

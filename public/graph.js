let data = null;

function loadGraph() {
    $.ajax({
        dataType: "json",
        type: "GET",
        url: "/api",
        success: function (json) {
            data = JSON.parse(atob(json));
            console.log(data);
            load_jsMind(data);
            $("#json-text-container").val(JSON.stringify(data));
            buildLinks();
        },
        error: function (error) {
            alert(error);
        },
    });
}

function load_jsMind(mind) {
    const options = {
        container: 'jsmind_container',
        editable: false,
        theme: 'belizehole'
    };
    let jm = new jsMind(options);
    // show it
    jm.show(mind);
}

function buildLinks() {
    $("jmnode").each(function () {
        const id = $(this).attr("nodeid");

        $($(this)).on("click", function () {
            $.ajax({
                dataType: "json",
                type: "GET",
                url: "/yaml_links",
                data: {
                    requestedId: id
                },
                success: function (json) {
                    const jsonPretty = JSON.stringify(JSON.parse(atob(json)), undefined, 2);
                    const cleanedJson = syntaxHighlight(jsonPretty);
                    $("#modalLongTitle").html(id);
                    $("#jsonData").html(cleanedJson);
                    $("#exampleModalLong").modal("toggle");
                },
                error: function (error) {
                    alert(error);
                },
            });
        });
    });
}

function syntaxHighlight(json) {
    if (typeof json != 'string') {
        json = JSON.stringify(json, undefined, 2);
    }
    json = json.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, function (match) {
        var cls = 'number';
        if (/^"/.test(match)) {
            if (/:$/.test(match)) {
                cls = 'key';
            } else {
                cls = 'string';
            }
        } else if (/true|false/.test(match)) {
            cls = 'boolean';
        } else if (/null/.test(match)) {
            cls = 'null';
        }

        if (cls === 'string') {
            return '<span class="' + cls + '">' + match + '</span>';
        } else {
            return '<span class="' + cls + '">' + match + '</span>';
        }
    });
}

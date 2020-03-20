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
        },
        error: function (error) {
            alert(error);
        },
    });
}

function load_jsMind(mind) {
    var options = {
        container: 'jsmind_container',
        editable: true,
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

// Old mxGraph function, probably won't be used.
function main(container) {
    // Checks if the browser is supported
    if (!mxClient.isBrowserSupported()) {
        mxUtils.error('Browser is not supported!', 200, false);
    } else {
        // Creates the graph inside the given container
        var graph = new mxGraph(container);

        // Enables rubberband selection
        new mxRubberband(graph);

        // Gets the default root for inserting new cells. This
        // is normally the first child of the root (ie. layer 0).
        var parent = graph.getDefaultParent();
        var root = undefined;

        // Adds cells to the model in a single step
        graph.getModel().beginUpdate();

        try {
            var dict = {};
            // run through each element in json
            data.forEach(function (element) {
                var name = element.name;
                // create graph element
                var graphElement = graph.insertVertex(parent, null,
                    name, 20, 20, 150, 30);
                // check if any parent element
                if (element.parentObjects.length > 0) {
                    // run through each parent element
                    element.parentObjects.forEach(function (parentObj) {
                        var parentGraphElement = dict[parentObj.name];
                        // add line between current element and parent
                        graph.insertEdge(parent, null, '', parentGraphElement, graphElement);
                    });
                } else {
                    // set root for layouting
                    root = graphElement;
                }
                // add element to dictionary. this is needed to find object later(parent)
                dict[name] = graphElement;
            });
        } finally {
            // Updates the display
            graph.getModel().endUpdate();

            // Creates a layout algorithm to be used
            // with the graph
            var layout = new mxStackLayout(graph, true);

            // Moves stuff wider apart than usual
            layout.forceConstant = 140;
            if (root) {
                layout.execute(parent, root);
            }
        }
    }
}

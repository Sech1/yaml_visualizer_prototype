let data = null;

function loadGraph() {
    $.ajax({
        dataType: "json",
        type: "GET",
        url: "/api",
        success: function (json) {
            data = atob(json);
            main($("#graphContainer"));
        },
        error: function (error) {
            alert(error);
        },
    });
}

// Program starts here. Creates a sample graph in the
// DOM node with the specified ID. This function is invoked
// from the onLoad event handler of the document (see below).
function main(container)
{
    // Checks if the browser is supported
    if (!mxClient.isBrowserSupported())
    {
        mxUtils.error('Browser is not supported!', 200, false);
    }
    else
    {
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

        try
        {
            var dict = {};
            // run through each element in json
            data.forEach(function(element) {
                var name = element.name;
                // create graph element
                var graphElement = graph.insertVertex(parent, null,
                    name, 20, 20, 80, 30);
                // check if any parent element
                if(element.parentObjects.length > 0) {
                    // run through each parent element
                    element.parentObjects.forEach(function(parentObj) {
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
        }
        finally
        {
            // Updates the display
            graph.getModel().endUpdate();

            // Creates a layout algorithm to be used
            // with the graph
            var layout = new mxHierarchicalLayout(graph, mxConstants.DIRECTION_NORTH);

            // Moves stuff wider apart than usual
            layout.forceConstant = 140;
            if(root) {
                layout.execute(parent, root);
            }
        }
    }
}
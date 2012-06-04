function(head, req) {
    // !json templates.head
    // !json templates.hook
    // !json templates.form
    // !json templates.tail

    provides("html", function() {
        var row;

        var data = {
            title: "Hooks",
            mainid: 'hooklist'
        };

        var Mustache = require("vendor/couchapp/lib/mustache");
        var path = require("vendor/couchapp/lib/path").init(req);

        send(Mustache.to_html(templates.head, data));
        send("<h1>List of Hooks</h1><ul id='hooks'>");

        while(row = getRow()) {
            send(Mustache.to_html(templates.hook, {
                id: row.id,
                rev: row.doc._rev,
                trigger: row.doc.trigger,
                url: row.doc.url,
                target: row.doc.target,
                show: path.show('hook', row.id)
            }));
        }

        send("</ul>");
        send(templates.form);
        send(Mustache.to_html(templates.tail, data));
    });
}
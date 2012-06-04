function initHookForm(app, t) {
    var form = $(t);
    form.submit(function() {
        var fdoc = form.serializeObject();
        fdoc.created_at = new Date();

        fdoc.events = [];
        $('#hookevents :checked').each(function() {
            fdoc.events.push($(this).val());
        });

        app.db.saveDoc(fdoc, {
            success : function() {
                form[0].reset();
                window.location.reload();
            }
        });
        return false;
    });
}

function initDeletes(app) {
    $("#hooks li.hook .trash").click(function(el) {
        var docrev = el.srcElement.parentElement.id.split(" ");
        console.log("Deleting", docrev);
        app.db.removeDoc({_id:docrev[0], _rev:docrev[1]}, {
            success: function() {
                $(el.srcElement.parentElement).hide(750);
            }
        });
    });

}
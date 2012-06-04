function(doc) {
    if (doc.type === 'hook') {
        emit([doc.trigger, doc.target], null);
    }
}
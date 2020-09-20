function addClass(element, className) {
    if (element.classList) {
        element.classList.add(className);
    } else {
        var classes = element.className.split(" ");

        classes.push(className);
        element.className = classes.join(" ");
    }
}

function removeClass(element, className) {
    if (element.classList) {
        element.classList.remove(className);
    } else {
        var classes = element.className.split(" ");
        var i = classes.indexOf(className);

        classes.splice(i, 1);
    }
}
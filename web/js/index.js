function search() {
    var s = document.getElementById("subjects");
    var subject = s.options[s.selectedIndex].value;

    var q = document.getElementById("question")
    var question = q.value

    var xhr = new XMLHttpRequest();
    xhr.onreadystatechange = function(response) {
        if (xhr.readyState === 4) {
            console.log(xhr.responseText)
        }
    };
    xhr.open('GET', '/finder?q=' + question + '&s=' + subject);
    xhr.send()
}
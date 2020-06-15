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
    xhr.open('GET', '/finder?q=' + question + '&s=' + subject,true);
    xhr.onload  = function() {
        var jsonResponse = req.response;
        
     };
    xhr.send()

}
function onload() {
    var xhr = new XMLHttpRequest();
    xhr.open('GET', '/subjects',true);
    
    xhr.onload  = function() {
        var jsonResponse =  JSON.parse(xhr.responseText);
        var subj = jsonResponse.Subjects.split(",")
        subj.forEach(addcat)
     };
     xhr.send()
    


}
function addcat(item,index) {
    var s = document.getElementById("subjects");
    var opt = document.createElement('option')
    opt.value = item;
    opt.innerHTML = item;
    s.appendChild(opt);


}
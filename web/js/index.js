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
    var jsonResponse =  JSON.parse(xhr.responseText);
    document.getElementById("query").innerHTML = jsonResponse.Query
    if (jsonResponse.Found == 'True') {
        document.getElementById("Paper").innerHTML = jsonResponse.Paper
        document.getElementById("qpl").innerHTML = jsonResponse.QPL
        document.getElementById("qpl").href = jsonResponse.QPL
        document.getElementById("msl").innerHTML = jsonResponse.MSL
    } else {
        document.getElementById("Paper").innerHTML = 'in no papers'
    }
    xhr.send()
 };

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
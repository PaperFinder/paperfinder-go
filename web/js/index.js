function search() {
    onload();

    var s = document.getElementById("subjects");

    var subject = s.options[s.selectedIndex].value;

    var q = document.getElementById("question");
    var question = q.value;

    var xhr = new XMLHttpRequest();

    xhr.onreadystatechange = function() {
        if (xhr.readyState === 4) {
            console.log(xhr.responseText)
        }
    };

    xhr.open('GET', '/finder?q=' + question + '&s=' + subject,false);
    
    xhr.send()
        var jsonResponse = JSON.parse(xhr.responseText);
        document.getElementById("result").innerHTML = gabi_content(jsonResponse.Query);
        if (jsonResponse.Found == 'True' || jsonResponse.Found == 'Partial') {
            if (jsonResponse.Found != 'Partial'){
                resultField.innerText = "\"" + gabi_content(jsonResponse.Query) + "\" was potentially found in "+ jsonResponse.Paper;
            } else {
                resultField.innerText = "\"" + gabi_content(jsonResponse.Query) + "\" was  found in "+ jsonResponse.Paper;
            }
            qplButton.href = jsonResponse.QPL;
            mslButton.href = jsonResponse.MSL;
    
            resultField.style.display = '';
            qplButton.style.display = '';
            mslButton.style.display = '';
        } else {
            resultField.innerHTML = "\"" + gabi_content(jsonResponse.Query) + "\" was found in no papers";
        }
 };

function gabi_content(content) {
    var temp = document.createElement("div");
    temp.innerHTML = content;
    
    return temp.textContent || temp.innerText || "";
}

function onload() {
    document.getElementById("result").style.display = 'block';
    document.getElementById("qpl").style.display = 'none';
    document.getElementById("msl").style.display = 'none';

    var xhr = new XMLHttpRequest();
    xhr.open('GET', '/subjects',true);
    
    xhr.onload  = function() {
        var jsonResponse =  JSON.parse(xhr.responseText);
        var subj = jsonResponse.Subjects.split(",")
        subj.forEach(addcat)
     };

     xhr.send()
}

function addcat(item, index) {
    var s = document.getElementById("subjects");
    var opt = document.createElement('option');

    opt.value = item;
    opt.innerHTML = item;

    if (s.childElementCount > index) {
        if (s.childElementCount[index].innerHTML == value) {
            return;
        }
    }

    s.appendChild(opt);
}
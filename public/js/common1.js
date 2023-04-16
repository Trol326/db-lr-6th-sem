var CityClassifier = function(nNumVariety, sIdDOM) {
    var formData = new FormData();
    formData.append('numvariety', nNumVariety.toString());
    var xhr = new XMLHttpRequest();
    xhr.open("POST", "/cityclassifier");
    xhr.onload = function(e)
    {
    if(this.readyState == 4 && this.status == 200)
    {
    var sAux = this.response;
    document.getElementById(sIdDOM).innerHTML = sAux;
    return false;
    }
    else
    {
    alert("Err!");
    }
    return false;
    };
    xhr.send(formData);
    return false;
    };
{{ define "nav.js" }}
function togglePostForm() {
    let pf = document.getElementById("section-submitForm").style.display;
    if (pf != "block") {
        document.getElementById("section-submitForm").style.display = "block";
        document.getElementById("newPostButt").innerHTML = "-";
        document.getElementById("newPostButt").style.background = "#8d561f";
    } else {
        document.getElementById("section-submitForm").style.display = "none";
        document.getElementById("newPostButt").innerHTML = "Post";
        document.getElementById("newPostButt").style.background = "#709624";
    }

}
function getStream(sortOrder) {
    window.location = window.location.origin + "/" + sortOrder;
}
{{end}}

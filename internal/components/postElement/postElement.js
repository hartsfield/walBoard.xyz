{{ define "postElement.js" }}
function collapsePost(postID) {
    if (document.getElementById("collapsable_" + postID).style.display == "block") {
        document.getElementById("collapseButt_" + postID).innerHTML = "[+]";
        document.getElementById("collapsable_" + postID).style.display = "none";
    } else {
        document.getElementById("collapseButt_" + postID).innerHTML = "[-]";
        document.getElementById("collapsable_" + postID).style.display = "block";
    }
}
{{end}}

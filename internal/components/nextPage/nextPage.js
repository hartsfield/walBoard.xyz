{{ define "nextPage.js" }}
let nextpagerButt = document.getElementById("nextPage");

let requestMade = false;
document.addEventListener("scroll", () => {
    if (isElementInViewport(nextpagerButt) && !requestMade) {
        requestMade = true;
        setTimeout(() => {
            submitNext();
        }, 500);
    }
});

let count = 20;
let lastPage = false;
async function submitNext() {
    if (!lastPage) {
        let postsWrapper = document.getElementById("postsWrapper")
        const response = await fetch("/{{.Order}}?count=" + count, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: "",
        });

        let res = await response.json();
        if (res.success == "true") {
            postsWrapper.insertAdjacentHTML("beforeend", res.template);
            requestMade = false;
            if (res.count != "None") {
                count = parseInt(res.count);
                console.log(count);
            } else {
                lastPage = true;
                document.getElementById("nextPage").innerHTML = "no more posts";
                document.getElementById("nextPage").style.fontSize = "1em";
                document.getElementById("nextPage").style.animationIterationCount = "1";
            }
        } else {
            console.log("error");
        }
    }
}
{{end}}

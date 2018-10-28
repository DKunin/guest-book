const form = document.querySelector('form');
form.addEventListener('submit', async (e) => {
    e.preventDefault();
    const name = e.target.elements.name.value;
    const text = e.target.elements.text.value;

    // const rawResponse = await fetch('/', {
    //     method: 'POST',
    //     headers: {
    //         'Accept': 'application/json',
    //         'Content-Type': 'application/x-www-form-urlencoded'
    //     },
    //     body: JSON.stringify({name, text})
    // });
    await fetch("/",
        {"credentials":"omit","headers":{
            "accept":"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
                "content-type":"application/x-www-form-urlencoded","upgrade-insecure-requests":"1"},
            "body":`name=${name}&text=${text}`,"method":"POST","mode":"cors"});
    const json = fetch()
    window.location.reload()
});

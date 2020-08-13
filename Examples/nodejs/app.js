const axios = require('axios')

let host = 'http://localhost:8080/cache';
axios.post(host + '/Hello', "World")
    .then(() => {
        axios.get(host + '/Hello')
            .then((response) => {
                console.log("Hello = ", response.data)
            })
    })
    .catch((error) => {
        console.error(error)
    });

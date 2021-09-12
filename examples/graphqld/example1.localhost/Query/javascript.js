#!/usr/bin/nodejs

let args = process.argv;
if (args[2] === '--graphqld-fields') {
    console.log(JSON.stringify(['javascript: String']));
} else {
    console.log(
        JSON.stringify(
            {
                msg: 'hello from javascript',
            }
        )
    )
}
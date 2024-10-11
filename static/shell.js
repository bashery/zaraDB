
//
const queryInput = document.getElementById('query-input');
const textarea = document.querySelector("textarea");

// load  pretty json setting
var pretty = localStorage.getItem("pretty") === "true" ? true : false;

// check and set pretty setteng ?
$(document).ready(function() {
   if (localStorage.getItem('pretty') == "true") {
       $('#togglePretty').prop('checked', true);
       pretty = true 
       return
   }

    $('#togglePretty').prop('checked', false);
    pretty = false

    setHightTextArea(textarea)
});

// change json's output format 
$('#togglePretty').change(function() {
    pretty = !pretty;
    
    if (pretty) {
       localStorage.setItem('pretty', "true");
    } else {
       localStorage.setItem('pretty', "false");
    }

    $('textarea').focus()
    sendQuery(queryInput)
    return;
});

// handle textarea events
queryInput.addEventListener('keydown', function(event) {

    // handle & send query
    if ((event.metaKey || event.altKey ) && event.key === 'Enter' ) {
        console.log("with Enter: ", queryInput)
        sendQuery(queryInput)
        return
    }

    // handle size of textarea
    if (event.key === 'Enter') {
        const cursorPosition = queryInput.selectionStart;

        // Insert a newline character at the cursor position
        const textBeforeCursor = queryInput.value.substring(0, cursorPosition);
        const textAfterCursor = queryInput.value.substring(cursorPosition);
        queryInput.value = textBeforeCursor + '\n' + textAfterCursor;

        // Move the cursor to the end of the newline
        queryInput.selectionStart = cursorPosition + 1;
        queryInput.selectionEnd = cursorPosition + 1;
        return 
    }
})

// render response data 
function HandleResponse(response) {
    $('#examples').hide();
    $('#data').html("<div><div>");
    if (pretty) {
        let Data = prettyJSON(response)
        $('#data').html(`<pre><span>${Data}</span></pre>`);
        $('#data').fadeIn(400);
        return;
    }

    data = JSON.parse(response);

    if (!Array.isArray(data)) {
        let obj = JSON.stringify(data) 
        $('#data').append(`<pre><span>${obj}</span></pre>`);
        $('#data').fadeIn(400);
        return;
    }

    for (let i = 0;i< data.length;i++) {
        let obj = JSON.stringify(data[i]) 
        obj = obj.replace(/,"/g, ', "'); 
        $('#data').append(`<pre><span>${obj}</span></pre>`);
    }
    $('#data').fadeIn(400);
};

// sendQuery db queries
function sendQuery(input) {

    $("#data").css("display","none");

   if (input.value.length< 5) {
        let error = "empty query"
        $('#data').html(`<pre><span>${error}</span></pre>`);
        $('#data').fadeIn(400);
        return;
    }
    //event.preventDefault();
    if (input.value) {
        try {
            eval("obj = "+ input.value)
            let query = JSON.stringify(obj)
            send(query)
            return;

        } catch (error) {
            console.error("Error evaluating code:", error);
            $('#data').html(`<pre><span>${error}</span></pre>`);
            $('#data').fadeIn(400);
            return; 
        }
    } 
}

// post query using HTTP POST request
function send(query) {
    $.ajax({
        url: 'http://localhost:1111/queries', // Replace with your actual server endpoint
        type: 'POST',
        contentType: 'text/plain', // Assuming your server accepts plain text
        data: query,
        success: function(response) {
            HandleResponse(response);
        },
        error: function(jqXHR, textStatus, errorThrown) {
            console.error("Error sending query:", errorThrown);
            $('#data').html(`<pre><span>${errorThrown}</span></pre>`);
            $('#data').fadeIn(400);
        }
    });
}

// pretty rendering json
function prettyJSON(jsonString) {
     try {
        const jsonObject = JSON.parse(jsonString);
        let res = JSON.stringify(jsonObject, null, 3);
        return  res
     } catch (error) {
        console.error("invalid json")
        return jsonString;
    }
}

// Dealing with Textarea Height
function calcHeight(value) {

    let numberOfLineBreaks = (value.match(/\n/g) || []).length;
    if (numberOfLineBreaks < 3) {
       numberOfLineBreaks = 3
    }
    
    if (numberOfLineBreaks > 24) {
        numberOfLineBreaks = 24
    }

    return 20 + numberOfLineBreaks * 20 + 12;
}

// reHight textarea
textarea.addEventListener("keyup", (e) => {
   // resize hight of textarea
   setHightTextArea(textarea)
});


// resize hight of textarea dynamicly
function setHightTextArea(textarea) {
    let hi = calcHeight(textarea.value) 
    textarea.style.height = hi + "px";

    hi = calcHeight(textarea.value)  + 10 
    $("#output").css("height", "calc(100vh - "+ hi +"px)" )
    //css height: calc(100vh - 100px);
}

// copy paste example into textarea
$('pre').click(function () {
    $('textarea').val($(this).text())
    $('textarea').focus()
})

// pointer on pre examples
$('pre').css('cursor', 'pointer');

//end

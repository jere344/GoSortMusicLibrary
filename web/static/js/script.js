document.addEventListener("DOMContentLoaded", function() {
    const form = document.getElementById("sortForm");
    const logContainer = document.getElementById("logContainer");
    const executionLogs = document.getElementById("executionLogs"); // Get the pre element
    const scriptTextarea = document.getElementById("script"); // Get the script textarea
    const scriptFileInput = document.getElementById("scriptFile"); // Get the file input

    // Handle file input change
    scriptFileInput.addEventListener("change", function(event) {
        const file = event.target.files[0];
        if (file) {
            const reader = new FileReader();
            reader.onload = function(e) {
                scriptTextarea.value = e.target.result; // Load file content into textarea
            };
            reader.onerror = function(e) {
                console.error("Error reading file:", e);
                alert("Error reading file: " + e.target.error);
            };
            reader.readAsText(file); // Read the file as text
        }
    });

    form.addEventListener("submit", function(event) {
        event.preventDefault();
        const formData = new FormData(form);
        // No need for Object.fromEntries if sending JSON directly
        const params = {
            sourceFolder: formData.get('sourceFolder'),
            destinationFolder: formData.get('destinationFolder'),
            script: formData.get('script'),
            fileOperationMode: formData.get('fileOperationMode') // Add file operation mode
        };

        fetch("/sort", {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify(params)
        })
        .then(response => {
            if (!response.ok) {
                return response.text().then(text => { throw new Error(text) });
            }
            return response.json();
         })
        .then(data => {
            displayLogs(data.logs || ["Operation completed successfully."]); // Display logs or a success message
        })
        .catch(error => {
            console.error("Error:", error);
            displayLogs([`Error: ${error.message}`]); // Display error message in the log container
        });
    });

    function displayLogs(logs) {
        executionLogs.textContent = ""; // Clear previous logs
        logs.forEach(log => {
            executionLogs.textContent += log + "\n"; // Append each log message
        });
    }
});
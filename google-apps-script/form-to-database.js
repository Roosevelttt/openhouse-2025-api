/**
 * Google Apps Script for connecting Google Forms to MySQL Database
 * 
 * Setup Instructions:
 * 1. Open your Google Form
 * 2. Click on "Responses" tab
 * 3. Click the Google Sheets icon to create a linked spreadsheet
 * 4. In the spreadsheet, go to Extensions > Apps Script
 * 5. Replace the default code with this script
 * 6. Save and run the createTrigger function once
 * 7. Update the API_ENDPOINT to your production server URL
 */

// Configuration - Update this to your production server URL
const API_ENDPOINT = 'https://your-production-server.com/api/user/update-angket';
// For local testing: 'http://localhost:8080/api/user/update-angket'

/**
 * Function that runs when a form is submitted
 * This function is automatically triggered when someone submits your Google Form
 */
function onFormSubmit(e) {
  try {
    // Get the submitted values
    const values = e.values;
    
    // Assuming your form structure is:
    // Column 0: Timestamp
    // Column 1: Name  
    // Column 2: NRP
    // Adjust these indices based on your actual form structure
    const timestamp = values[0];
    const name = values[1];      // Adjust index if needed
    const nrp = values[2];       // Adjust index if needed
    
    console.log('Form submitted:', { timestamp, name, nrp });
    
    // Validate required fields
    if (!name || !nrp) {
      console.error('Missing required fields:', { name, nrp });
      return;
    }
    
    // Send data to your API
    updateDatabaseAngket(nrp, name);
    
  } catch (error) {
    console.error('Error in onFormSubmit:', error);
  }
}

/**
 * Function to update database via API
 */
function updateDatabaseAngket(nrp, name) {
  const payload = {
    'nrp': nrp,
    'name': name
  };
  
  const options = {
    'method': 'POST',
    'headers': {
      'Content-Type': 'application/json',
    },
    'payload': JSON.stringify(payload)
  };
  
  try {
    console.log('Sending to API:', payload);
    
    const response = UrlFetchApp.fetch(API_ENDPOINT, options);
    const responseText = response.getContentText();
    const statusCode = response.getResponseCode();
    
    console.log('API Response:', {
      statusCode: statusCode,
      response: responseText
    });
    
    if (statusCode === 200) {
      console.log('Successfully updated angket status for NRP:', nrp);
    } else {
      console.error('API Error:', responseText);
    }
    
  } catch (error) {
    console.error('Error calling API:', error);
  }
}

/**
 * Function to create the form submit trigger
 * Run this function ONCE to set up the automatic trigger
 */
function createTrigger() {
  try {
    // Get the linked form
    const sheet = SpreadsheetApp.getActiveSheet();
    const form = FormApp.openByUrl(sheet.getFormUrl());
    
    // Create trigger for form submissions
    ScriptApp.newTrigger('onFormSubmit')
      .event(ScriptApp.EventType.ON_FORM_SUBMIT)
      .create();
      
    console.log('Trigger created successfully!');
    
  } catch (error) {
    console.error('Error creating trigger:', error);
    console.log('Make sure this spreadsheet is linked to a Google Form');
  }
}

/**
 * Function to test the API connection
 * Run this function to test if your API is working
 */
function testAPIConnection() {
  console.log('Testing API connection...');
  
  // Test with sample data
  updateDatabaseAngket('123456789', 'Test User');
}

/**
 * Function to process existing responses (bulk update)
 * Run this if you want to update existing form responses
 */
function processExistingResponses() {
  try {
    const sheet = SpreadsheetApp.getActiveSheet();
    const data = sheet.getDataRange().getValues();
    
    // Skip header row (index 0)
    for (let i = 1; i < data.length; i++) {
      const row = data[i];
      const name = row[1];  // Adjust index based on your form
      const nrp = row[2];   // Adjust index based on your form
      
      if (name && nrp) {
        console.log(`Processing row ${i + 1}: ${name} (${nrp})`);
        updateDatabaseAngket(nrp, name);
        
        // Add small delay to avoid rate limiting
        Utilities.sleep(500);
      }
    }
    
    console.log('Finished processing existing responses');
    
  } catch (error) {
    console.error('Error processing existing responses:', error);
  }
}

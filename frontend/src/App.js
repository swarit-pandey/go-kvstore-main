// src/App.js

import React from 'react';
import axios from 'axios';
import AppLayout from './components/AppLayout';
import './App.css';

const App = () => {
  const [response, setResponse] = React.useState({ status: null, message: null });

  const handleCommandSubmit = async (command) => {
    try {
      const res = await axios.post('http://localhost:8080/api/commands', {
        command: command,
      });
      console.log('API Response:', res);
      setResponse({ status: res.status, message: res.data });
    } catch (error) {
      setResponse({
        status: error.response.status,
        message: 'An error occurred while processing the command.',
      });
    }
  };

  return (
    <div>
      <div className="layout-container">
        <AppLayout onCommandSubmit={handleCommandSubmit} />
        {
          response.message && (
            <div className="response-message">
              {response.status === 200 ? (
                <p>[Success] Response: {response.message}</p>
              ) : (
                <p>[Error] Response: {response.message ? response.message : response}</p>
              )}
            </div>
          )
        }
      </div>
    </div>
  );
}

export default App;

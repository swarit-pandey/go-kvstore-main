// src/App.js

import React from 'react';
import axios from 'axios';
import AppLayout from './components/AppLayout';
import './App.css';

const App = () => {
  const [response, setResponse] = React.useState({ status: null, message: null });
  const [countdown, setCountdown] = React.useState(null);
  const [recentKeys, setRecentKeys] = React.useState([]);

  const handleCommandSubmit = async (command) => {
    try {
      const res = await axios.post('http://localhost:8080/api/commands', {
        command: command,
      });
      console.log('API Response:', res);
      setResponse({ status: res.status, message: res.data });

      // Start the countdown if the command has an expiration time
      const commandParts = command.trim().split(' ');
      if (commandParts[0].toLowerCase() === 'set' && commandParts.length === 5 && commandParts[3].toLowerCase() === 'ex') {
        const expirationTime = parseInt(commandParts[4]);
        setCountdown(expirationTime);
      }
    } catch (error) {
      setResponse({
        status: error.response.status,
        message: 'An error occurred while processing the command.',
      });
    }

    const commandParts = command.trim().split(' ');
    if (commandParts[0].toLowerCase() === 'set' && commandParts.length >= 3) {
      const key = commandParts[1];
      const value = commandParts[2];
      setRecentKeys((prevKeys) => [...prevKeys, { key, value }]);
    }
  };



  React.useEffect(() => {
    if (countdown === null || countdown <= 0) {
      return;
    }

    const timerId = setTimeout(() => {
      setCountdown(countdown - 1);
    }, 1000);

    return () => {
      clearTimeout(timerId);
    };
  }, [countdown]);

  return (
    <div>
      <AppLayout onCommandSubmit={handleCommandSubmit} />
      {response && (
        <div className="response-message">
          {response.status === 200 ? (
            <p>[Success] Response: {response.message}</p>
          ) : (
            <p>[Error] Response: {response.message}</p>
          )}
        </div>
      )}
      {countdown !== null && countdown > 0 && (
        <div className="countdown-timer">
          <p>Key will expire in {countdown} seconds</p>
        </div>
      )}
      <div className="recent-keys-table">
        <table>
          <thead>
            <tr>
              <th>Key</th>
              <th>Value</th>
            </tr>
          </thead>
          <tbody>
            {recentKeys.map((keyValuePair, index) => (
              <tr key={index}>
                <td>{keyValuePair.key}</td>
                <td>{keyValuePair.value}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );


}

export default App;

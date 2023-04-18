import React from 'react';
import { Container, Box, TextField, Button, Typography } from '@mui/material';
import { Send } from '@mui/icons-material';

const AppLayout = ({ onCommandSubmit }) => { // Change this line
    const [command, setCommand] = React.useState('');

    const handleSubmit = (e) => {
        e.preventDefault();
        onCommandSubmit(command);
        setCommand('');
    };

    return (
        <Container>
            <Box
                component="form"
                onSubmit={handleSubmit}
                sx={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    flexDirection: 'column',
                    minHeight: '100vh',
                }}
            >
                <Typography variant="h4" component="h1" gutterBottom>
                    Go-KVStore
                </Typography>
                <TextField
                    label="Enter Command"
                    variant="outlined"
                    value={command}
                    onChange={(e) => setCommand(e.target.value)}
                    sx={{ width: '100%', maxWidth: 400 }}
                />
                <Box sx={{ mt: 2 }}>
                    <Button
                        type="submit"
                        variant="contained"
                        color="primary"
                        endIcon={<Send />}
                    >
                        Send Command
                    </Button>
                </Box>
            </Box>
        </Container>
    );
};

export default AppLayout;
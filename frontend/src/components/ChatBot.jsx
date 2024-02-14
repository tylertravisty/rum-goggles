import { useEffect, useState } from 'react';
import { Modal, SmallModal } from './Modal';
import { LoginChatBot, UpdateChatBotUrl } from '../../wailsjs/go/main/App';

import './ChatBot.css';

export function ChatBotModal(props) {
    const [error, setError] = useState('');
    const [loggedIn, setLoggedIn] = useState(props.loggedIn);
    const [password, setPassword] = useState('');
    const [saving, setSaving] = useState(false);
    const updatePassword = (event) => setPassword(event.target.value);
    const [url, setUrl] = useState(props.streamUrl);
    const updateUrl = (event) => setUrl(event.target.value);
    const [username, setUsername] = useState(props.username);
    const updateUsername = (event) => setUsername(event.target.value);

    useEffect(() => {
        if (saving) {
            // let user = username;
            // let p = password;
            // let u = url;
            // props.onSubmit(user, p, u);
            // NewChatBot(props.cid, username, password, url)
            if (loggedIn) {
                UpdateChatBotUrl(props.cid, url)
                    .then(() => {
                        reset();
                        props.onUpdate(url);
                    })
                    .catch((error) => {
                        setSaving(false);
                        setError(error);
                        console.log('Error updating chat bot:', error);
                    });
            } else {
                LoginChatBot(props.cid, username, password, url)
                    .then(() => {
                        reset();
                        props.onLogin();
                    })
                    .catch((error) => {
                        setSaving(false);
                        setError(error);
                        console.log('Error creating new chat bot:', error);
                    });
            }
        }
    }, [saving]);

    const reset = () => {
        setError('');
        setLoggedIn(false);
        setPassword('');
        setSaving(false);
        setUrl('');
        setUsername('');
    };

    const close = () => {
        reset();
        props.onClose();
    };

    const logout = () => {
        reset();
        props.onLogout();
    };

    const submit = () => {
        if (username === '') {
            setError('Add username');
            return;
        }

        if (password === '' && !loggedIn) {
            setError('Add password');
            return;
        }

        if (url === '') {
            setError('Add stream URL');
            return;
        }

        setSaving(true);
        // let user = username;
        // let p = password;
        // let u = url;
        // reset();
        // props.onSubmit(user, p, u);
    };

    return (
        <>
            <Modal
                onClose={close}
                show={props.show}
                style={{ minWidth: '300px', maxWidth: '400px' }}
                cancelButton={loggedIn ? '' : 'Cancel'}
                onCancel={close}
                deleteButton={loggedIn ? 'Logout' : ''}
                onDelete={logout}
                submitButton={saving ? 'Saving' : 'Save'}
                onSubmit={
                    saving
                        ? () => {
                              console.log('Saving');
                          }
                        : submit
                }
                title={'Chat Bot'}
            >
                <div className='chat-bot-modal'>
                    {loggedIn ? (
                        <div className='chat-bot-description'>
                            <span className='chat-bot-description-label'>Logged in:</span>
                            <span
                                className='chat-bot-description-label'
                                style={{ fontWeight: 'bold' }}
                            >
                                {username}
                            </span>
                        </div>
                    ) : (
                        <div className='chat-bot-setting'>
                            <span className='chat-bot-setting-label'>Username</span>
                            <input
                                className='chat-bot-setting-input'
                                onChange={updateUsername}
                                placeholder='Username'
                                type='text'
                                value={username}
                            />
                        </div>
                    )}
                    {!loggedIn && (
                        <div className='chat-bot-setting'>
                            <span className='chat-bot-setting-label'>Password</span>
                            <input
                                className='chat-bot-setting-input'
                                onChange={updatePassword}
                                placeholder='Password'
                                type='password'
                                value={password}
                            />
                        </div>
                    )}
                    <div className='chat-bot-setting'>
                        <span className='chat-bot-setting-label'>Stream URL</span>
                        <input
                            className='chat-bot-setting-input'
                            onChange={updateUrl}
                            placeholder='https://'
                            type='text'
                            value={url}
                        />
                    </div>
                </div>
            </Modal>
            <SmallModal
                onClose={() => setError('')}
                show={error !== ''}
                style={{ minWidth: '300px', maxWidth: '300px', maxHeight: '100px' }}
                title={'Error'}
                message={error}
                submitButton={'OK'}
                onSubmit={() => setError('')}
            />
        </>
    );
}

export function StreamChatMessageItem() {}

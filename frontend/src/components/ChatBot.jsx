import { useEffect, useState } from 'react';

import { Modal, SmallModal } from './Modal';

import { NewChatBot } from '../../wailsjs/go/main/App';

import './ChatBot.css';

export function ChatBotModal(props) {
    const [error, setError] = useState('');
    const [password, setPassword] = useState('');
    const [saving, setSaving] = useState(false);
    const updatePassword = (event) => setPassword(event.target.value);
    const [url, setUrl] = useState('');
    const updateUrl = (event) => setUrl(event.target.value);
    const [username, setUsername] = useState('');
    const updateUsername = (event) => setUsername(event.target.value);

    useEffect(() => {
        if (saving) {
            // let user = username;
            // let p = password;
            // let u = url;
            // props.onSubmit(user, p, u);
            NewChatBot(props.cid, username, password, url)
                .then(() => {
                    reset();
                    props.onClose();
                })
                .catch((error) => {
                    setSaving(false);
                    setError(error);
                    console.log('Error creating new chat bot:', error);
                });
        }
    }, [saving]);

    const reset = () => {
        setError('');
        setPassword('');
        setSaving(false);
        setUrl('');
        setUsername('');
    };

    const close = () => {
        reset();
        props.onClose();
    };

    const submit = () => {
        if (username === '') {
            setError('Add username');
            return;
        }

        if (password === '') {
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
                cancelButton={'Cancel'}
                onCancel={close}
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
                    {/* {error && <span className='chat-bot-error'>{error}</span>} */}
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

import { useEffect, useState } from 'react';

import { Modal, SmallModal } from './Modal';

import { OpenFileDialog } from '../../wailsjs/go/main/App';

import './StreamChatMessage.css';

export function StreamChatMessageModal(props) {
    const [asChannel, setAsChannel] = useState(props.asChannel);
    const [chatCommand, setChatCommand] = useState(props.chatCommand);
    const [error, setError] = useState('');
    const [onCommand, setOnCommand] = useState(props.onCommand);
    const [openDelete, setOpenDelete] = useState(false);
    const [readFromFile, setReadFromFile] = useState(false);
    const [text, setText] = useState(props.text);
    const [textFile, setTextFile] = useState(props.textFile);
    const updateText = (event) => setText(event.target.value);
    const [timer, setTimer] = useState(props.interval);

    useEffect(() => {
        console.log('update chat');
        setAsChannel(props.asChannel);
        setOnCommand(props.onCommand);
        setError('');
        setReadFromFile(props.textFile !== '');
        setText(props.text);
        setTextFile(props.textFile);
        setTimer(props.interval);
    }, []);

    const reset = () => {
        setAsChannel(false);
        setChatCommand(false);
        setError('');
        setReadFromFile(false);
        setText('');
        setTextFile('');
        setOnCommand(false);
        setTimer('');
    };

    const close = () => {
        reset();
        props.onClose();
    };

    const submit = () => {
        if (!readFromFile && text === '') {
            setError('Add message');
            return;
        }

        if (readFromFile && textFile === '') {
            setError('Select file containing messages');
            return;
        }

        if (timer === '') {
            setError('Set timer');
            return;
        }

        if (onCommand && chatCommand === '') {
            setError('Add command');
            return;
        }

        let ac = asChannel;
        let oc = onCommand;
        let cmd = chatCommand;
        let int = timerToInterval();
        let txt = text;
        let txtfile = textFile;
        reset();
        props.onSubmit(props.chatID, ac, cmd, int, oc, txt, txtfile);
    };

    const deleteMessage = () => {
        if (props.chatID === '') {
            close();
            return;
        }

        setOpenDelete(true);
    };

    const confirmDelete = () => {
        reset();
        setOpenDelete(false);
        props.onDelete(props.chatID);
    };

    const updateChatCommand = (e) => {
        let command = e.target.value;

        if (command.length === 1) {
            if (command !== '!') {
                command = '!' + command;
            }
        }
        command = command.toLowerCase();
        let postfix = command.replace('!', '');

        if (postfix !== '' && !/^[a-z0-9]+$/gi.test(postfix)) {
            return;
        }

        setChatCommand(command);
    };

    const updateTimerBackspace = (e) => {
        if (timer.length === 0) {
            return;
        }

        if (e.keyCode === 8) {
            setTimer(timer.substring(0, timer.length - 1));
        }
    };

    const updateTimer = (e) => {
        let nums = '0123456789';
        let digit = e.target.value;

        if (!nums.includes(digit)) {
            return;
        }

        if (timer.length === 6) {
            return;
        }

        if (timer.length === 0 && digit === '0') {
            return;
        }

        setTimer(timer + digit);
    };

    const timerToInterval = () => {
        let prefix = '0'.repeat(6 - timer.length);
        let t = prefix + timer;

        let hours = parseInt(t.substring(0, 2));
        let minutes = parseInt(t.substring(2, 4));
        let seconds = parseInt(t.substring(4, 6));

        return hours * 3600 + minutes * 60 + seconds;
    };

    const printTimer = () => {
        if (timer === '') {
            return '00:00:00';
        }

        let prefix = '0'.repeat(6 - timer.length);
        let t = prefix + timer;

        return t.substring(0, 2) + ':' + t.substring(2, 4) + ':' + t.substring(4, 6);
    };

    const checkChannelToggle = (e) => {
        setAsChannel(e.target.checked);
    };

    const checkCommandToggle = (e) => {
        setOnCommand(e.target.checked);
    };

    const checkReadFromFile = (e) => {
        setReadFromFile(e.target.checked);
        if (!e.target.checked) {
            setTextFile('');
        }
    };

    const chooseFile = () => {
        OpenFileDialog()
            .then((filepath) => {
                if (filepath !== '') {
                    setTextFile(filepath);
                }
            })
            .catch((error) => setError(error));
    };

    return (
        <>
            <Modal
                onClose={close}
                show={props.show}
                style={{ minWidth: '300px', maxWidth: '400px' }}
                cancelButton={props.chatID === '' ? 'Cancel' : ''}
                onCancel={deleteMessage}
                deleteButton={props.chatID === '' ? '' : 'Delete'}
                onDelete={deleteMessage}
                style={{ minHeight: '450px', maxWidth: '400px' }}
                submitButton={'Save'}
                onSubmit={submit}
                title={'Chat Message'}
            >
                <div className='stream-chat-message-modal'>
                    <div className='stream-chat-message'>
                        {/* {error && <span className='stream-chat-message-error'>{error}</span>} */}
                        <div className='stream-chat-message-title'>
                            <span className='stream-chat-message-label'>Message</span>
                            <div className='stream-chat-message-title-right'>
                                <span className='chat-toggle-check-label'>Read from file</span>
                                <label className='chat-toggle-check-container'>
                                    <input
                                        checked={readFromFile}
                                        onChange={checkReadFromFile}
                                        type='checkbox'
                                    />
                                    <span className='chat-toggle-check'></span>
                                </label>
                            </div>
                        </div>
                        {readFromFile ? (
                            <div className='choose-file'>
                                <div className='choose-file-button-box'>
                                    <button className='choose-file-button' onClick={chooseFile}>
                                        Choose file
                                    </button>
                                </div>
                                <span className='choose-file-path'>{textFile}</span>
                            </div>
                        ) : (
                            <textarea
                                className='stream-chat-message-textarea'
                                cols='25'
                                onChange={updateText}
                                rows='4'
                                value={text}
                            />
                        )}
                    </div>
                    <div className='chat-options'>
                        <div className='chat-interval'>
                            <span className='chat-interval-label'>
                                {onCommand ? 'Command timeout' : 'Chat interval'}
                            </span>
                            <input
                                className={
                                    timer === ''
                                        ? 'chat-interval-input chat-interval-input-zero'
                                        : 'chat-interval-input chat-interval-input-value'
                                }
                                onKeyDown={updateTimerBackspace}
                                onInput={updateTimer}
                                placeholder={printTimer()}
                                size='8'
                                type='text'
                                value={''}
                            />
                        </div>
                        <div className='chat-toggle'>
                            <span className='chat-toggle-label'>Chat as channel</span>
                            <label className='chat-toggle-switch'>
                                <input
                                    onChange={checkChannelToggle}
                                    type='checkbox'
                                    checked={asChannel}
                                />
                                <span className='chat-toggle-slider round'></span>
                            </label>
                        </div>
                        <div className='chat-toggle'>
                            <span className='chat-toggle-label'>Chat on command</span>
                            <label className='chat-toggle-switch'>
                                <input
                                    onChange={checkCommandToggle}
                                    type='checkbox'
                                    checked={onCommand}
                                />
                                <span className='chat-toggle-slider round'></span>
                            </label>
                        </div>
                        {onCommand ? (
                            <div className='chat-command'>
                                <input
                                    className='chat-command-input'
                                    onInput={updateChatCommand}
                                    placeholder={'!command'}
                                    size='8'
                                    type='text'
                                    value={chatCommand}
                                />
                            </div>
                        ) : (
                            <div className='chat-command'>
                                <span className='chat-command-label'>{'\u00A0'}</span>
                            </div>
                        )}
                    </div>
                </div>
            </Modal>
            <SmallModal
                onClose={() => setOpenDelete(false)}
                show={openDelete}
                style={{ minWidth: '300px', maxWidth: '200px', maxHeight: '200px' }}
                cancelButton={'Cancel'}
                onCancel={() => setOpenDelete(false)}
                deleteButton={'Delete'}
                message={
                    'Are you sure you want to delete this message? You cannot undo this action.'
                }
                onDelete={confirmDelete}
                title={'Delete Message'}
            />
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

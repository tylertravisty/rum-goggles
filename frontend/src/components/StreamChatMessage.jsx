import { useEffect, useState } from 'react';

import { Modal, SmallModal } from './Modal';

import './StreamChatMessage.css';

export function StreamChatMessageModal(props) {
    const [asChannel, setAsChannel] = useState(props.asChannel);
    const [openDelete, setOpenDelete] = useState(false);
    const [error, setError] = useState('');
    const [message, setMessage] = useState(props.message);
    const updateMessage = (event) => setMessage(event.target.value);
    const [timer, setTimer] = useState(props.interval);

    useEffect(() => {
        console.log('update chat');
        setAsChannel(props.asChannel);
        setError('');
        setMessage(props.message);
        setTimer(props.interval);
    }, []);

    const reset = () => {
        setAsChannel(false);
        setError('');
        setMessage('');
        setTimer('');
    };

    const close = () => {
        reset();
        props.onClose();
    };

    const submit = () => {
        if (message === '') {
            setError('Add message');
            return;
        }

        if (timer === '') {
            setError('Set timer');
            return;
        }

        let ac = asChannel;
        let msg = message;
        let int = timerToInterval();
        reset();
        props.onSubmit(props.chatID, ac, int, msg);
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

    const checkToggle = (e) => {
        setAsChannel(e.target.checked);
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
                submitButton={'Save'}
                onSubmit={submit}
                title={'Chat Message'}
            >
                <div className='stream-chat-message-modal'>
                    <div className='stream-chat-message'>
                        {error && <span className='stream-chat-message-error'>{error}</span>}
                        <div className='stream-chat-message-title'>
                            <span className='stream-chat-message-label'>Message</span>
                        </div>
                        <textarea
                            className='stream-chat-message-textarea'
                            cols='25'
                            onChange={updateMessage}
                            rows='4'
                            value={message}
                        />
                    </div>
                    <div className='chat-options'>
                        <div className='chat-interval'>
                            <span className='chat-interval-label'>Chat interval</span>
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
                        <div className='chat-as-channel'>
                            <span className='chat-as-channel-label'>Chat as channel</span>
                            <label className='chat-as-channel-switch'>
                                <input onChange={checkToggle} type='checkbox' checked={asChannel} />
                                <span className='chat-as-channel-slider round'></span>
                            </label>
                        </div>
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
        </>
    );
}

export function StreamChatMessageItem() {}

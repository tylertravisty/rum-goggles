import { useEffect, useState } from 'react';
import { Modal, SmallModal } from './Modal';
import {
    AccountList,
    ChatbotList,
    DeleteChatbot,
    NewChatbot,
    OpenFileDialog,
    UpdateChatbot,
} from '../../wailsjs/go/main/App';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import { ChevronLeft, ChevronRight, Gear, PlusCircle, Robot } from '../assets';
import './ChatBot.css';

function ChatBot(props) {
    const [chatbots, setChatbots] = useState([]);
    const [deleteChatbot, setDeleteChatbot] = useState(false);
    const [editChatbot, setEditChatbot] = useState(null);
    const [error, setError] = useState('');
    const [openChatbot, setOpenChatbot] = useState(null);
    const [openNewChatbot, setOpenNewChatbot] = useState(false);
    const [openNewRule, setOpenNewRule] = useState(false);
    const [chatbotSettings, setChatbotSettings] = useState(true);

    useEffect(() => {
        EventsOn('ChatbotList', (event) => {
            setChatbots(event);
            if (openChatbot !== null) {
                for (const chatbot of event) {
                    if (chatbot.id === openChatbot.id) {
                        setOpenChatbot(chatbot);
                    }
                }
            }
        });
    }, []);

    useEffect(() => {
        ChatbotList()
            .then((response) => {
                setChatbots(response);
            })
            .catch((error) => {
                setError(error);
            });
    }, []);

    const open = (chatbot) => {
        setOpenChatbot(chatbot);
    };

    const closeEdit = () => {
        setEditChatbot(null);
    };

    const openEdit = (chatbot) => {
        setEditChatbot(chatbot);
    };

    const openNew = () => {
        setOpenNewChatbot(true);
    };

    const sortChatbots = () => {
        let sorted = [...chatbots].sort((a, b) =>
            a.name.toLowerCase() > b.name.toLowerCase() ? 1 : -1
        );

        return sorted;
    };

    const confirmDelete = () => {
        DeleteChatbot(openChatbot)
            .then(() => {
                setDeleteChatbot(false);
                setEditChatbot(null);
                setOpenChatbot(null);
            })
            .catch((err) => {
                setDeleteChatbot(false);
                setError(err);
            });
    };

    return (
        <>
            {error !== '' && (
                <SmallModal
                    onClose={() => setError('')}
                    show={error !== ''}
                    style={{ minWidth: '300px', maxWidth: '200px', maxHeight: '200px' }}
                    title={'Error'}
                    message={error}
                    submitButton={'OK'}
                    onSubmit={() => setError('')}
                />
            )}
            {openNewChatbot && (
                <ModalChatbot
                    cancelButton={'Cancel'}
                    onClose={() => setOpenNewChatbot(false)}
                    show={setOpenNewChatbot}
                    submit={NewChatbot}
                    submitButton={'Create'}
                    submittingButton={'Creating...'}
                    title={'New Chatbot'}
                />
            )}
            {editChatbot !== null && (
                <ModalChatbot
                    chatbot={editChatbot}
                    onClose={closeEdit}
                    deleteButton={'Delete'}
                    onDelete={() => setDeleteChatbot(true)}
                    show={editChatbot !== null}
                    submit={UpdateChatbot}
                    submitButton={'Update'}
                    submittingButton={'Updating...'}
                    title={'Edit Chatbot'}
                />
            )}
            {deleteChatbot && (
                <SmallModal
                    cancelButton={'Cancel'}
                    onCancel={() => setDeleteChatbot(false)}
                    onClose={() => setDeleteChatbot(false)}
                    show={deleteChatbot}
                    style={{ minWidth: '300px', maxWidth: '200px', maxHeight: '200px' }}
                    title={'Delete Chatbot'}
                    message={
                        'Are you sure you want to delete the chatbot? All rules associated with this chatbot will be deleted as well.'
                    }
                    submitButton={'OK'}
                    onSubmit={confirmDelete}
                />
            )}
            {openNewRule && (
                <ModalNewRule onClose={() => setOpenNewRule(false)} show={openNewRule} />
            )}
            <div className='chatbot'>
                {openChatbot === null ? (
                    <>
                        <div className='chatbot-header'>
                            <div className='chatbot-header-left'>
                                <img className='chatbot-header-icon' src={Robot} />
                                <span className='chatbot-header-title'>Bots</span>
                            </div>
                            <div className='chatbot-header-right'>
                                <button className='chatbot-header-button' onClick={openNew}>
                                    <img className='chatbot-header-button-icon' src={PlusCircle} />
                                </button>
                            </div>
                        </div>
                        <div className='chatbot-list'>
                            {sortChatbots().map((chatbot, index) => (
                                <ChatbotListItem chatbot={chatbot} key={index} onClick={open} />
                            ))}
                        </div>
                    </>
                ) : (
                    <div className='chatbot-header'>
                        <div className='chatbot-header-left'>
                            <img
                                className='chatbot-header-icon-back'
                                onClick={() => setOpenChatbot(null)}
                                src={ChevronLeft}
                            />
                        </div>
                        <span className='chatbot-header-title'>{openChatbot.name}</span>
                        <div className='chatbot-header-right'>
                            <button
                                className='chatbot-header-button'
                                onClick={() => openEdit(openChatbot)}
                            >
                                <img className='chatbot-header-button-icon' src={Gear} />
                            </button>
                            <button
                                className='chatbot-header-button'
                                onClick={() => setOpenNewRule(true)}
                            >
                                <img className='chatbot-header-button-icon' src={PlusCircle} />
                            </button>
                        </div>
                    </div>
                )}
            </div>
        </>
    );
}

export default ChatBot;

function ChatbotListItem(props) {
    return (
        <div className='chatbot-list-item'>
            <button className='chatbot-list-button' onClick={() => props.onClick(props.chatbot)}>
                <span className='chatbot-list-item-name'>{props.chatbot.name}</span>
            </button>
        </div>
    );
}

function ModalChatbot(props) {
    const [error, setError] = useState('');
    const [id, setId] = useState(props.chatbot === undefined ? null : props.chatbot.id);
    const [loading, setLoading] = useState(false);
    const [name, setName] = useState(props.chatbot === undefined ? '' : props.chatbot.name);
    const updateName = (event) => {
        if (loading) {
            return;
        }
        setName(event.target.value);
    };
    const [nameValid, setNameValid] = useState(true);
    const [url, setUrl] = useState(props.chatbot === undefined ? '' : props.chatbot.url);
    const updateUrl = (event) => {
        if (loading) {
            return;
        }
        setUrl(event.target.value);
    };

    useEffect(() => {
        if (loading) {
            props
                .submit({ id: id, name: name, url: url })
                .then(() => {
                    reset();
                    props.onClose();
                })
                .catch((err) => {
                    setLoading(false);
                    setError(err);
                });
        }
    }, [loading]);

    const close = () => {
        if (loading) {
            return;
        }

        reset();
        props.onClose();
    };

    const reset = () => {
        setLoading(false);
        setName('');
        setNameValid(true);
    };

    const submit = () => {
        if (name == '') {
            setNameValid(false);
            return;
        }

        setLoading(true);
    };

    return (
        <>
            {error !== '' && (
                <SmallModal
                    onClose={() => setError('')}
                    show={error !== ''}
                    style={{ minWidth: '300px', maxWidth: '200px', maxHeight: '200px' }}
                    title={'Error'}
                    message={error}
                    submitButton={'OK'}
                    onSubmit={() => setError('')}
                />
            )}
            <Modal
                cancelButton={props.cancelButton}
                onCancel={close}
                onClose={close}
                deleteActive={true}
                deleteButton={props.deleteButton}
                onDelete={props.onDelete}
                show={props.show}
                style={{ minWidth: '400px', maxWidth: '400px', maxHeight: '400px' }}
                submitButton={loading ? props.submittingButton : props.submitButton}
                onSubmit={submit}
                title={props.title}
            >
                <div className='chatbot-modal-form'>
                    {nameValid ? (
                        <label className='chatbot-modal-label'>Chatbot Name</label>
                    ) : (
                        <label className='chatbot-modal-label-warning'>
                            Chatbot Name - Please enter a valid name
                        </label>
                    )}
                    <input
                        className='chatbot-modal-input'
                        onChange={updateName}
                        placeholder={'Name'}
                        type={'text'}
                        value={name}
                    ></input>
                    <label className='chatbot-modal-label'>Live Stream URL</label>
                    <input
                        className='chatbot-modal-input'
                        onChange={updateUrl}
                        placeholder={'https://rumble.com'}
                        type={'text'}
                        value={url}
                    ></input>
                </div>
            </Modal>
        </>
    );
}

function ModalNewRule(props) {
    const [back, setBack] = useState([]);
    const [rule, setRule] = useState({});
    const [stage, setStage] = useState('trigger');
    const updateStage = (next, reverse) => {
        setBack([...back, { stage: stage, reverse: reverse }]);
        setStage(next);
    };

    const goBack = () => {
        if (back.length === 0) {
            return;
        }

        const last = back.at(-1);
        setStage(last.stage);
        if (last.reverse !== undefined && last.reverse !== null) {
            setRule(last.reverse(rule));
        }
        setBack(back.slice(0, back.length - 1));
    };

    const submit = () => {};

    return (
        <>
            {stage === 'trigger' && (
                <ModalNewRuleTrigger
                    onClose={props.onClose}
                    rule={rule}
                    setRule={setRule}
                    setStage={updateStage}
                    show={props.show}
                />
            )}
            {stage === 'trigger-timer' && (
                <ModalNewRuleTriggerTimer
                    onBack={goBack}
                    onClose={props.onClose}
                    rule={rule}
                    setRule={setRule}
                    setStage={updateStage}
                    show={props.show}
                />
            )}
            {stage === 'message' && (
                <ModalNewRuleMessage
                    onBack={goBack}
                    onClose={props.onClose}
                    rule={rule}
                    setRule={setRule}
                    setStage={updateStage}
                    show={props.show}
                />
            )}
            {stage === 'sender' && (
                <ModalNewRuleSender
                    onBack={goBack}
                    onClose={props.onClose}
                    rule={rule}
                    setRule={setRule}
                    setStage={updateStage}
                    show={props.show}
                />
            )}
            {stage === 'review' && (
                <ModalNewRuleReview
                    onBack={goBack}
                    onClose={props.onClose}
                    rule={rule}
                    show={props.show}
                    onSubmit={submit}
                />
            )}
        </>
    );
}

function ModalNewRuleTrigger(props) {
    const next = (stage) => {
        const rule = props.rule;
        rule.trigger = {};
        props.setRule(rule);
        props.setStage(stage, reverse);
    };

    const reverse = (rule) => {
        rule.trigger = null;
        return rule;
    };

    return (
        <Modal
            onClose={props.onClose}
            show={props.show}
            style={{ height: '480px', minHeight: '480px', width: '360px', minWidth: '360px' }}
        >
            <div className='modal-add-account-channel'>
                <span className='modal-add-account-channel-title'>Choose Rule Trigger</span>
                <div className='modal-add-account-channel-body'>
                    <button
                        className='modal-add-account-channel-button'
                        onClick={() => next('trigger-command')}
                    >
                        <div className='modal-add-account-channel-button-left'>
                            <span>Command</span>
                        </div>
                        <img
                            className='modal-add-account-channel-button-right-icon'
                            src={ChevronRight}
                        />
                    </button>
                    <button
                        className='modal-add-account-channel-button'
                        onClick={() => next('trigger-stream_event')}
                    >
                        <div className='modal-add-account-channel-button-left'>
                            <span>Stream Event</span>
                        </div>
                        <img
                            className='modal-add-account-channel-button-right-icon'
                            src={ChevronRight}
                        />
                    </button>
                    <button
                        className='modal-add-account-channel-button'
                        onClick={() => next('trigger-timer')}
                    >
                        <div className='modal-add-account-channel-button-left'>
                            <span>Timer</span>
                        </div>
                        <img
                            className='modal-add-account-channel-button-right-icon'
                            src={ChevronRight}
                        />
                    </button>
                </div>
                <div></div>
            </div>
        </Modal>
    );
}

function ModalNewRuleTriggerTimer(props) {
    const [validTimer, setValidTimer] = useState(true);
    const [timer, setTimer] = useState(
        props.rule.trigger.on_timer !== undefined && props.rule.trigger.on_timer !== null
            ? props.rule.trigger.on_timer
            : ''
    );

    const back = () => {
        const rule = props.rule;
        rule.trigger.on_timer = '';
        props.setRule(rule);
        props.onBack();
    };

    const next = () => {
        if (timer === '') {
            setValidTimer(false);
            return;
        }

        const rule = props.rule;
        rule.trigger.on_timer = timer;
        props.setRule(rule);
        props.setStage('message', null);
    };

    return (
        <Modal
            cancelButton={'Back'}
            onCancel={back}
            onClose={props.onClose}
            show={props.show}
            submitButton={'Next'}
            onSubmit={next}
            style={{ height: '300px', minHeight: '300px', width: '360px', minWidth: '360px' }}
        >
            <div className='modal-add-account-channel'>
                <div className='modal-add-account-channel-header'>
                    <span className='modal-add-account-channel-title'>Set Timer</span>
                    <span className='modal-add-account-channel-subtitle'>
                        Chat rule will trigger at the set interval.
                    </span>
                </div>
                <div className='modal-add-account-channel-body'>
                    {validTimer ? (
                        <span className='chatbot-modal-description'>Enter timer</span>
                    ) : (
                        <span className='chatbot-modal-description-warning'>
                            Enter a valid timer interval.
                        </span>
                    )}
                    <Timer timer={timer} setTimer={setTimer} />
                </div>
                <div style={{ height: '56px' }}></div>
            </div>
        </Modal>
    );
}

function ModalNewRuleMessage(props) {
    const [error, setError] = useState('');
    const [message, setMessage] = useState(
        props.rule.message !== undefined && props.rule.message !== null ? props.rule.message : {}
    );
    const [refresh, setRefresh] = useState(false);
    const [validFile, setValidFile] = useState(true);
    const [validText, setValidText] = useState(true);

    const back = () => {
        const rule = props.rule;
        rule.message = null;
        props.setRule(rule);
        props.onBack();
    };

    const next = () => {
        if (fromFile()) {
            if (message.from_file.filepath === undefined || message.from_file.filepath === '') {
                setValidFile(false);
                return;
            }
        } else {
            if (message.from_text === undefined || message.from_text === '') {
                setValidText(false);
                return;
            }
        }

        const rule = props.rule;
        rule.message = message;
        props.setRule(rule);
        props.setStage('sender', null);
    };

    const chooseFile = () => {
        OpenFileDialog()
            .then((filepath) => {
                if (filepath !== '') {
                    message.from_file.filepath = filepath;
                    setMessage(message);
                    setRefresh(!refresh);
                }
            })
            .catch((error) => setError(error));
    };

    const fromFile = () => {
        return message.from_file !== undefined && message.from_file !== null;
    };

    const toggleFromFile = () => {
        if (fromFile()) {
            message.from_file = null;
        } else {
            message.from_file = {};
        }

        setMessage(message);
        setRefresh(!refresh);
    };

    const randomRead = () => {
        if (!fromFile()) {
            return false;
        }

        if (message.from_file.random_read === undefined || message.from_file.random_read === null) {
            return false;
        }

        return message.from_file.random_read;
    };

    const toggleRandomRead = () => {
        if (!fromFile()) {
            return;
        }

        message.from_file.random_read = !randomRead();
        setMessage(message);
        setRefresh(!refresh);
    };

    const updateMessageText = (event) => {
        message.from_text = event.target.value;
        setMessage(message);
    };

    const updateMessageFilepath = (filepath) => {
        if (!fromFile()) {
            message.from_file = {};
        }
        message.from_file = filepath;
        setMessage(message);
    };

    return (
        <>
            {error !== '' && (
                <SmallModal
                    onClose={() => setError('')}
                    show={error !== ''}
                    style={{ minWidth: '300px', maxWidth: '200px', maxHeight: '200px' }}
                    title={'Error'}
                    message={error}
                    submitButton={'OK'}
                    onSubmit={() => setError('')}
                />
            )}
            <Modal
                cancelButton={'Back'}
                onCancel={back}
                onClose={props.onClose}
                show={props.show}
                submitButton={'Next'}
                onSubmit={next}
                style={{ height: '480px', minHeight: '480px', width: '360px', minWidth: '360px' }}
            >
                <div className='modal-add-account-channel'>
                    <span className='modal-add-account-channel-title'>Add Message</span>
                    <div className='modal-add-account-channel-body'>
                        {fromFile() ? (
                            validFile ? (
                                <label className='modal-add-channel-label'>Message</label>
                            ) : (
                                <label className='modal-add-channel-label-warning'>
                                    Select a file
                                </label>
                            )
                        ) : validText ? (
                            <label className='modal-add-channel-label'>Message</label>
                        ) : (
                            <label className='modal-add-channel-label-warning'>Add a message</label>
                        )}
                        {fromFile() ? (
                            <div className='choose-file'>
                                <div className='choose-file-button-box'>
                                    <button className='choose-file-button' onClick={chooseFile}>
                                        Choose file
                                    </button>
                                </div>
                                <span className='choose-file-path'>
                                    {message.from_file.filepath}
                                </span>
                            </div>
                        ) : (
                            <textarea
                                className='chatbot-modal-textarea'
                                onChange={updateMessageText}
                                rows='4'
                                value={message.from_text}
                            />
                        )}
                        <div className='chatbot-modal-setting'>
                            <label className='chatbot-modal-setting-description'>
                                Read from file
                            </label>
                            <label className='chatbot-modal-toggle-switch'>
                                <input
                                    checked={fromFile()}
                                    onChange={toggleFromFile}
                                    type='checkbox'
                                />
                                <span className='chatbot-modal-toggle-slider round'></span>
                            </label>
                        </div>
                        {fromFile() && (
                            <div className='chatbot-modal-setting'>
                                <label className='chatbot-modal-setting-description'>
                                    Choose lines in random order
                                </label>
                                <label className='chatbot-modal-toggle-switch'>
                                    <input
                                        checked={randomRead()}
                                        onChange={toggleRandomRead}
                                        type='checkbox'
                                    />
                                    <span className='chatbot-modal-toggle-slider round'></span>
                                </label>
                            </div>
                        )}
                    </div>
                    <div style={{ height: '29px' }}></div>
                </div>
            </Modal>
        </>
    );
}

function ModalNewRuleSender(props) {
    const [accounts, setAccounts] = useState({});
    const [error, setError] = useState('');
    const [sender, setSender] = useState(
        props.rule.send_as !== undefined && props.rule.send_as !== null ? props.rule.send_as : {}
    );
    const [validSender, setValidSender] = useState(true);

    useEffect(() => {
        AccountList()
            .then((response) => {
                setAccounts(response);
            })
            .catch((error) => {
                setError(error);
            });
    }, []);

    const back = () => {
        const rule = props.rule;
        rule.send_as = null;
        props.setRule(rule);
        props.onBack();
    };

    const next = () => {
        if (sender.username === undefined || sender.username === '') {
            setValidSender(false);
            return;
        }

        const rule = props.rule;
        rule.send_as = sender;
        props.setRule(rule);
        props.setStage('review', null);
    };

    const selectSender = (sender) => {
        setSender(sender);
    };

    const sortChannels = (channels) => {
        let sorted = [...channels].sort((a, b) =>
            a.name.toLowerCase() > b.name.toLowerCase() ? 1 : -1
        );

        return sorted;
    };

    const sortAccounts = () => {
        let keys = Object.keys(accounts);

        let sorted = [...keys].sort((a, b) =>
            accounts[a].account.username.toLowerCase() > accounts[b].account.username.toLowerCase()
                ? 1
                : -1
        );

        return sorted;
    };

    const sortPages = () => {
        let pages = [];

        const keys = sortAccounts();
        keys.forEach((key, i) => {
            const account = accounts[key];
            pages.push({
                display: account.account.username,
                channel_id: null,
                username: account.account.username,
            });
            const channels = sortChannels(account.channels);
            channels.forEach((channel, j) => {
                pages.push({
                    display: channel.name,
                    channel_id: channel.cid,
                    username: account.account.username,
                });
            });
        });

        return pages;
    };

    return (
        <>
            {error !== '' && (
                <SmallModal
                    onClose={() => setError('')}
                    show={error !== ''}
                    style={{ minWidth: '300px', maxWidth: '200px', maxHeight: '200px' }}
                    title={'Error'}
                    message={error}
                    submitButton={'OK'}
                    onSubmit={() => setError('')}
                />
            )}
            <Modal
                cancelButton={'Back'}
                onCancel={back}
                onClose={props.onClose}
                show={props.show}
                submitButton={'Next'}
                onSubmit={next}
                style={{ height: '480px', minHeight: '480px', width: '360px', minWidth: '360px' }}
            >
                <div className='modal-add-account-channel'>
                    <div className='modal-add-account-channel-header'>
                        <span className='modal-add-account-channel-title'>Choose sender</span>
                    </div>
                    <div className='modal-add-account-channel-body' style={{ height: '60%' }}>
                        {validSender ? (
                            <span className='chatbot-modal-description'>Chat As</span>
                        ) : (
                            <span className='chatbot-modal-description-warning'>
                                Select an account or channel
                            </span>
                        )}
                        <div className='chatbot-modal-pages'>
                            {sortPages().map((page, index) => (
                                <div className={'chatbot-modal-page'} key={index}>
                                    <button
                                        className='chatbot-modal-page-button'
                                        onClick={() => selectSender(page)}
                                        style={{
                                            backgroundColor:
                                                page.display === sender.display ? '#85c742' : '',
                                        }}
                                    >
                                        {page.display}
                                    </button>
                                </div>
                            ))}
                        </div>
                    </div>
                    <div style={{ height: '29px' }}></div>
                </div>
            </Modal>
        </>
    );
}

function ModalNewRuleReview(props) {
    const [error, setError] = useState('');

    const back = () => {
        props.onBack();
    };

    const submit = () => {
        props.onSubmit();
    };

    return (
        <Modal
            cancelButton={'Back'}
            onCancel={back}
            onClose={props.onClose}
            show={props.show}
            submitButton={'Submit'}
            onSubmit={submit}
            style={{ height: '480px', minHeight: '480px', width: '360px', minWidth: '360px' }}
        >
            <div className='modal-add-account-channel'>
                <div className='modal-add-account-channel-header'>
                    <span className='modal-add-account-channel-title'>Review</span>
                </div>
                <div className='modal-add-account-channel-body'>
                    <div className='chatbot-modal-setting'>
                        <label className='chatbot-modal-setting-description'>Trigger</label>
                        <label className='chatbot-modal-setting-description'>Timer</label>
                    </div>
                    <div className='chatbot-modal-setting'>
                        <label className='chatbot-modal-setting-description'>Chat As</label>
                        <label className='chatbot-modal-setting-description'>tylertravisty</label>
                    </div>
                </div>
                <div style={{ height: '29px' }}></div>
            </div>
        </Modal>
    );
}

function Timer(props) {
    const updateTimerBackspace = (e) => {
        if (props.timer.length === 0) {
            return;
        }

        if (e.keyCode === 8) {
            props.setTimer(props.timer.substring(0, props.timer.length - 1));
        }
    };

    const updateTimer = (e) => {
        let nums = '0123456789';
        let digit = e.target.value;

        if (!nums.includes(digit)) {
            return;
        }

        let interval = timerValToInterval(props.timer + digit);
        if (interval >= 360000) {
            return;
        }

        if (props.timer.length === 0 && digit === '0') {
            return;
        }

        props.setTimer(props.timer + digit);
    };

    const timerValToInterval = (val) => {
        let prefix = '0'.repeat(6 - val.length);
        let t = prefix + val;

        let hours = parseInt(t.substring(0, 2));
        let minutes = parseInt(t.substring(2, 4));
        let seconds = parseInt(t.substring(4, 6));

        return hours * 3600 + minutes * 60 + seconds;
    };

    const printTimer = () => {
        if (props.timer === '') {
            return '00:00:00';
        }

        let prefix = '0'.repeat(6 - props.timer.length);
        let t = prefix + props.timer;

        return t.substring(0, 2) + ':' + t.substring(2, 4) + ':' + t.substring(4, 6);
    };

    return (
        <input
            className={
                props.timer === ''
                    ? 'timer-input timer-input-zero'
                    : 'timer-input timer-input-value'
            }
            onKeyDown={updateTimerBackspace}
            onInput={updateTimer}
            placeholder={printTimer()}
            size='8'
            type='text'
            value={''}
        />
    );
}

import { useEffect, useState } from 'react';
import { Modal, SmallModal } from './Modal';
import {
    AccountList,
    ChatbotList,
    ChatbotRules,
    DeleteChatbot,
    DeleteChatbotRule,
    NewChatbot,
    NewChatbotRule,
    OpenFileDialog,
    RunChatbotRule,
    RunChatbotRules,
    StopChatbotRule,
    StopChatbotRules,
    UpdateChatbot,
    UpdateChatbotRule,
} from '../../wailsjs/go/main/App';
import { EventsOff, EventsOn } from '../../wailsjs/runtime/runtime';
import {
    ChevronLeft,
    ChevronRight,
    Gear,
    GearWhite,
    PauseBig,
    PlayBig,
    PlayBigGreen,
    PlusCircle,
    Robot,
    StopBigRed,
} from '../assets';
import './ChatBot.css';
import { DropDown } from './DropDown';

function ChatBot(props) {
    const [chatbots, setChatbots] = useState([]);
    const [deleteChatbot, setDeleteChatbot] = useState(false);
    const [editChatbot, setEditChatbot] = useState(null);
    const [error, setError] = useState('');
    const [openChatbot, setOpenChatbot] = useState(null);
    const [openNewChatbot, setOpenNewChatbot] = useState(false);
    const [openNewRule, setOpenNewRule] = useState(false);
    const [chatbotRules, setChatbotRules] = useState([]);
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

        EventsOn('ChatbotRules', (event) => {
            setChatbotRules(event);
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
        ChatbotRules(chatbot)
            .then((response) => {
                setChatbotRules(response);
                setOpenChatbot(chatbot);
            })
            .catch((error) => {
                setError(error);
            });
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

    const deleteChatbotRule = (rule) => {
        DeleteChatbotRule(rule).catch((error) => setError(error));
    };

    const startAll = () => {
        RunChatbotRules(openChatbot.id).catch((error) => {
            setError(error);
        });
    };

    const stopAll = () => {
        StopChatbotRules(openChatbot.id).catch((error) => {
            setError(error);
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
                <ModalRule
                    chatbot={openChatbot}
                    onClose={() => setOpenNewRule(false)}
                    new={true}
                    show={openNewRule}
                />
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
                    <>
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
                        <div className='chatbot-rule'>
                            <span className='chatbot-rule-header chatbot-rule-output'>Output</span>
                            <span className='chatbot-rule-header chatbot-rule-trigger'>
                                Trigger
                            </span>
                            <span className='chatbot-rule-header chatbot-rule-sender'>Sender</span>
                            <div className='chatbot-rule-buttons'>
                                <button className='chatbot-rule-button' onClick={startAll}>
                                    <img className='chatbot-rule-button-icon' src={PlayBigGreen} />
                                </button>
                                <button className='chatbot-rule-button' onClick={stopAll}>
                                    <img className='chatbot-rule-button-icon' src={StopBigRed} />
                                </button>
                            </div>
                        </div>
                        <div className='chatbot-rules'>
                            {chatbotRules.map((rule, index) => (
                                <ChatbotRule
                                    chatbot={openChatbot}
                                    deleteRule={deleteChatbotRule}
                                    rule={rule}
                                    key={index}
                                />
                            ))}
                        </div>
                    </>
                )}
            </div>
        </>
    );
}

export default ChatBot;

function ChatbotListItem(props) {
    return (
        <div className='chatbot-list-item'>
            <button
                className='chatbot-list-item-button'
                onClick={() => props.onClick(props.chatbot)}
            >
                <span className='chatbot-list-item-name'>{props.chatbot.name}</span>
            </button>
        </div>
    );
}

function ChatbotRule(props) {
    const [ruleActive, setRuleActive] = useState(props.rule.running);
    const updateRuleActive = (active) => {
        props.rule.running = active;
        setRuleActive(active);
    };
    const [ruleError, setRuleError] = useState('');
    const [ruleID, setRuleID] = useState(0);
    const [updateRule, setUpdateRule] = useState(false);

    useEffect(() => {
        if (ruleID !== props.rule.id) {
            EventsOff('ChatbotRuleActive-' + props.rule.id);
            EventsOff('ChatbotRuleError-' + props.rule.id);
        }

        EventsOn('ChatbotRuleActive-' + props.rule.id, (event) => {
            updateRuleActive(event);
        });

        EventsOn('ChatbotRuleError-' + props.rule.id, (event) => {
            setRuleError(event);
        });

        setRuleID(props.rule.id);
    }, [props.rule.id]);

    useEffect(() => {
        setRuleActive(props.rule.running);
    }, [props.rule.running]);

    const deleteRule = () => {
        setUpdateRule(false);
        props.deleteRule(props.rule);
    };

    const prettyTimer = (timer) => {
        let hours = Math.floor(timer / 3600);
        let minutes = Math.floor(timer / 60 - hours * 60);
        let seconds = Math.floor(timer - hours * 3600 - minutes * 60);

        return hours + 'h ' + minutes + 'm ' + seconds + 's';
    };

    const printTriggerEvent = () => {
        const onEvent = props.rule.parameters.trigger.on_event;
        switch (true) {
            case onEvent.from_account !== undefined && onEvent.from_account !== null:
                const fromAccount = props.rule.parameters.trigger.on_event.from_account;
                switch (true) {
                    case fromAccount.on_follow !== undefined && fromAccount.on_follow !== null:
                        return 'Follow';
                    default:
                        return '';
                }
                break;
            case onEvent.from_channel !== undefined && onEvent.from_channel !== null:
                const fromChannel = props.rule.parameters.trigger.on_event.from_channel;
                switch (true) {
                    case fromChannel.on_follow !== undefined && fromChannel.on_follow !== null:
                        return 'Follow';
                    default:
                        return '';
                }
                break;
            case onEvent.from_live_stream !== undefined && onEvent.from_live_stream !== null:
                const fromLiveStream = props.rule.parameters.trigger.on_event.from_live_stream;
                switch (true) {
                    case fromLiveStream.on_raid !== undefined && fromLiveStream.on_raid !== null:
                        return 'Raid';
                    case fromLiveStream.on_rant !== undefined && fromLiveStream.on_rant !== null:
                        return 'Rant';
                    case fromLiveStream.on_sub !== undefined && fromLiveStream.on_sub !== null:
                        return 'Sub';
                    default:
                        return '';
                }
            default:
                return '';
        }
    };

    const printTrigger = () => {
        let trigger = props.rule.parameters.trigger;

        switch (true) {
            case trigger.on_command !== undefined && trigger.on_command !== null:
                return trigger.on_command.command;
            case trigger.on_event !== undefined && trigger.on_event !== null:
                return printTriggerEvent();
            case trigger.on_timer !== undefined && trigger.on_timer !== null:
                return prettyTimer(props.rule.parameters.trigger.on_timer);
        }
    };

    const startRule = () => {
        setRuleActive(true);
        RunChatbotRule(props.rule)
            .then(() => {
                updateRuleActive(true);
            })
            .catch((error) => {
                // TODO: format error in rule with exclamation point indicator
                // Replace play/pause button with exclamation point
                // User must clear error before reactivating
                setRuleActive(false);
            });
    };

    const stopRule = () => {
        let active = ruleActive;
        setRuleActive(false);
        StopChatbotRule(props.rule)
            .then(() => {
                updateRuleActive(false);
            })
            .catch((error) => {
                setRuleActive(active);
                // TODO: format error in rule with exclamation point indicator
                // Replace play/pause button with exclamation point
                // User must clear error before reactivating
            });
    };

    const triggerKey = () => {
        const trigger = props.rule.parameters.trigger;

        switch (true) {
            case trigger.on_command !== undefined && trigger.on_command !== null:
                return 'on_command';
            case trigger.on_event !== undefined && trigger.on_event !== null:
                return 'on_event';
            case trigger.on_timer !== undefined && trigger.on_timer !== null:
                return 'on_timer';
        }
    };

    return (
        <>
            {updateRule && (
                <ModalRule
                    chatbot={props.chatbot}
                    onClose={() => setUpdateRule(false)}
                    onDelete={deleteRule}
                    new={false}
                    rule={JSON.parse(JSON.stringify(props.rule.parameters))}
                    ruleID={props.rule.id}
                    show={updateRule}
                    trigger={triggerKey()}
                />
            )}
            <div className='chatbot-rule'>
                <span className='chatbot-rule-output'>{props.rule.display}</span>
                <span className='chatbot-rule-trigger'>{printTrigger()}</span>
                <span className='chatbot-rule-sender'>{props.rule.parameters.send_as.display}</span>
                <div className='chatbot-rule-buttons'>
                    <button
                        className='chatbot-rule-button'
                        onClick={ruleActive ? stopRule : startRule}
                    >
                        <img
                            className='chatbot-rule-button-icon'
                            src={ruleActive ? PauseBig : PlayBig}
                        />
                    </button>
                    <button className='chatbot-rule-button' onClick={() => setUpdateRule(true)}>
                        <img className='chatbot-rule-button-icon' src={GearWhite} />
                    </button>
                </div>
            </div>
        </>
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

function ModalRule(props) {
    const [back, setBack] = useState(
        props.new
            ? []
            : [
                  { stage: 'trigger' },
                  { stage: 'trigger-' + props.trigger },
                  { stage: 'message' },
                  { stage: 'sender' },
              ]
    );
    const [edit, setEdit] = useState(props.new ? true : false);
    const [error, setError] = useState('');
    const [rule, setRule] = useState(props.new ? {} : props.rule);
    const [stage, setStage] = useState(props.new ? 'trigger' : 'review');
    const updateStage = (next) => {
        setBack([...back, { stage: stage }]);
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

    const submit = () => {
        if (props.new) {
            submitNew();
        }

        submitUpdate();
    };

    const submitNew = () => {
        let appRule = {
            chatbot_id: props.chatbot.id,
            parameters: rule,
        };

        NewChatbotRule(appRule)
            .then(() => {
                props.onClose();
            })
            .catch((err) => {
                setError(err);
            });
    };

    const submitUpdate = () => {
        let appRule = {
            id: props.ruleID,
            chatbot_id: props.chatbot.id,
            parameters: rule,
        };

        UpdateChatbotRule(appRule)
            .then(() => {
                props.onClose();
            })
            .catch((err) => {
                setError(err);
            });
    };

    console.log('back:', back);
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
            {stage === 'event-from_stream' && (
                <ModalRuleEventStream
                    onBack={goBack}
                    onClose={props.onClose}
                    rule={rule}
                    setRule={setRule}
                    setStage={updateStage}
                    show={props.show}
                />
            )}
            {stage === 'message' && (
                <ModalRuleMessage
                    onBack={goBack}
                    onClose={props.onClose}
                    rule={rule}
                    setRule={setRule}
                    setStage={updateStage}
                    show={props.show}
                />
            )}
            {stage === 'review' && (
                <ModalRuleReview
                    edit={edit}
                    setEdit={setEdit}
                    new={props.new}
                    onBack={goBack}
                    onClose={props.onClose}
                    onDelete={props.onDelete}
                    onSubmit={submit}
                    rule={rule}
                    show={props.show}
                />
            )}
            {stage === 'sender' && (
                <ModalRuleSender
                    onBack={goBack}
                    onClose={props.onClose}
                    rule={rule}
                    setRule={setRule}
                    setStage={updateStage}
                    show={props.show}
                />
            )}
            {stage === 'trigger' && (
                <ModalRuleTrigger
                    onClose={props.onClose}
                    rule={rule}
                    setRule={setRule}
                    setStage={updateStage}
                    show={props.show}
                />
            )}
            {stage === 'trigger-on_command' && (
                <ModalRuleTriggerCommand
                    onBack={goBack}
                    onClose={props.onClose}
                    rule={rule}
                    setRule={setRule}
                    setStage={updateStage}
                    show={props.show}
                />
            )}
            {stage === 'trigger-on_event' && (
                <ModalRuleTriggerEvent
                    onBack={goBack}
                    onClose={props.onClose}
                    rule={rule}
                    setRule={setRule}
                    setStage={updateStage}
                    show={props.show}
                />
            )}
            {stage === 'trigger-on_timer' && (
                <ModalRuleTriggerTimer
                    onBack={goBack}
                    onClose={props.onClose}
                    rule={rule}
                    setRule={setRule}
                    setStage={updateStage}
                    show={props.show}
                />
            )}
        </>
    );
}

function ModalRuleTrigger(props) {
    const next = (stage) => {
        props.setStage(stage);
    };

    const triggerOnCommand = () => {
        const rule = props.rule;
        if (rule.trigger == undefined || rule.trigger == null) {
            rule.trigger = {};
        }
        if (rule.trigger.on_command == undefined || rule.trigger.on_command == null) {
            rule.trigger.on_command = {};
        }
        if (
            rule.trigger.on_command.restrict == undefined ||
            rule.trigger.on_command.restrict == null
        ) {
            rule.trigger.on_command.restrict = {};
        }

        rule.trigger.on_event = null;
        rule.trigger.on_timer = null;

        props.setRule(rule);

        next('trigger-on_command');
    };

    const triggerOnEvent = () => {
        const rule = props.rule;
        if (rule.trigger == undefined || rule.trigger == null) {
            rule.trigger = {};
        }

        rule.trigger.on_command = null;
        rule.trigger.on_timer = null;

        props.setRule(rule);

        next('trigger-on_event');
    };

    const triggerOnTimer = () => {
        const rule = props.rule;
        if (rule.trigger == undefined || rule.trigger == null) {
            rule.trigger = {};
        }

        rule.trigger.on_command = null;
        rule.trigger.on_event = null;

        props.setRule(rule);

        next('trigger-on_timer');
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
                    <button className='modal-add-account-channel-button' onClick={triggerOnCommand}>
                        <div className='modal-add-account-channel-button-left'>
                            <span>Command</span>
                        </div>
                        <img
                            className='modal-add-account-channel-button-right-icon'
                            src={ChevronRight}
                        />
                    </button>
                    <button className='modal-add-account-channel-button' onClick={triggerOnEvent}>
                        <div className='modal-add-account-channel-button-left'>
                            <span>Event</span>
                        </div>
                        <img
                            className='modal-add-account-channel-button-right-icon'
                            src={ChevronRight}
                        />
                    </button>
                    <button className='modal-add-account-channel-button' onClick={triggerOnTimer}>
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

function ModalRuleTriggerCommand(props) {
    const prependZero = (value) => {
        if (value < 10) {
            return '0' + value;
        }

        return '' + value;
    };

    const intervalToTimer = (interval) => {
        let hours = Math.floor(interval / 3600);
        let minutes = Math.floor(interval / 60 - hours * 60);
        let seconds = Math.floor(interval - hours * 3600 - minutes * 60);

        if (hours !== 0 || minutes !== 0) {
            seconds = prependZero(seconds);
        }
        if (hours !== 0) {
            minutes = prependZero(minutes);
        }
        if (hours === 0) {
            hours = '';
            if (minutes === 0) {
                minutes = '';
                if (seconds === 0) {
                    seconds = '';
                }
            }
        }

        return hours + minutes + seconds;
    };

    const [validCommand, setValidCommand] = useState('');
    const [command, setCommand] = useState(
        props.rule.trigger.on_command.command !== undefined
            ? props.rule.trigger.on_command.command
            : ''
    );
    const [followersOnly, setFollowersOnly] = useState(
        props.rule.trigger.on_command.restrict.to_follower !== undefined
            ? props.rule.trigger.on_command.restrict.to_follower
            : false
    );
    const toggleFollowersOnly = () => {
        setFollowersOnly(!followersOnly);
    };
    const [minRant, setMinRant] = useState(
        props.rule.trigger.on_command.restrict.to_rant !== undefined
            ? props.rule.trigger.on_command.restrict.to_rant
            : 0
    );
    const updateMinRant = (e) => {
        let amount = parseInt(e.target.value);
        if (isNaN(amount)) {
            amount = 0;
        }

        setMinRant(amount);
    };
    const [subscribersOnly, setSubscribersOnly] = useState(
        props.rule.trigger.on_command.restrict.to_subscriber !== undefined
            ? props.rule.trigger.on_command.restrict.to_subscriber
            : false
    );
    const toggleSubscribersOnly = () => {
        setSubscribersOnly(!subscribersOnly);
    };
    const [timeout, setTimeout] = useState(
        props.rule.trigger.on_command.timeout !== undefined
            ? intervalToTimer(props.rule.trigger.on_command.timeout)
            : ''
    );

    const back = () => {
        // const rule = props.rule;
        // rule.trigger.on_command = null;
        // props.setRule(rule);
        props.onBack();
    };

    const next = () => {
        if (command === '') {
            setValidCommand('Enter a valid command');
            return;
        }
        if (timeout === '') {
            setValidCommand('Enter a valid timeout');
            return;
        }

        const rule = props.rule;
        rule.trigger.on_command = {
            command: command,
            restrict: {
                to_follower: followersOnly,
                to_rant: minRant,
                to_subscriber: subscribersOnly,
            },
            timeout: timerValToInterval(timeout),
        };
        props.setRule(rule);
        props.setStage('message');
    };

    const timerValToInterval = (val) => {
        let prefix = '0'.repeat(6 - val.length);
        let t = prefix + val;

        let hours = parseInt(t.substring(0, 2));
        let minutes = parseInt(t.substring(2, 4));
        let seconds = parseInt(t.substring(4, 6));

        return hours * 3600 + minutes * 60 + seconds;
    };

    return (
        <Modal
            cancelButton={'Back'}
            onCancel={back}
            onClose={props.onClose}
            show={props.show}
            submitButton={'Next'}
            onSubmit={next}
            style={{ height: '400px', minHeight: '300px', width: '360px', minWidth: '360px' }}
        >
            <div className='modal-add-account-channel'>
                <div className='modal-add-account-channel-header'>
                    <span className='modal-add-account-channel-title'>Set Command</span>
                    <span className='modal-add-account-channel-subtitle'>
                        Chat rule will trigger on command.
                    </span>
                </div>
                <div className='modal-add-account-channel-body'>
                    {validCommand === '' ? (
                        <span className='chatbot-modal-description'>Enter command</span>
                    ) : (
                        <span className='chatbot-modal-description-warning'>{validCommand}</span>
                    )}
                    <Command command={command} setCommand={setCommand} />
                    <div className='chatbot-modal-setting'>
                        <label className='chatbot-modal-setting-description'>Set Timeout</label>
                        <Timer timer={timeout} setTimer={setTimeout} />
                    </div>
                    <div className='chatbot-modal-setting'>
                        <label className='chatbot-modal-setting-description'>Followers only</label>
                        <label className='chatbot-modal-toggle-switch'>
                            <input
                                checked={followersOnly}
                                onChange={toggleFollowersOnly}
                                type='checkbox'
                            />
                            <span className='chatbot-modal-toggle-slider round'></span>
                        </label>
                    </div>
                    <div className='chatbot-modal-setting'>
                        <label className='chatbot-modal-setting-description'>
                            Subscribers only
                        </label>
                        <label className='chatbot-modal-toggle-switch'>
                            <input
                                checked={subscribersOnly}
                                onChange={toggleSubscribersOnly}
                                type='checkbox'
                            />
                            <span className='chatbot-modal-toggle-slider round'></span>
                        </label>
                    </div>
                    <div className='chatbot-modal-setting'>
                        <label className='chatbot-modal-setting-description'>
                            Minimum rant amount
                        </label>
                        <div>
                            <span className='command-rant-amount-symbol'>$</span>
                            <input
                                className='command-rant-amount'
                                onChange={updateMinRant}
                                placeholder='0'
                                size='4'
                                type='text'
                                value={minRant === 0 ? '' : minRant}
                            />
                        </div>
                    </div>
                </div>
                <div style={{ height: '56px' }}></div>
            </div>
        </Modal>
    );
}

function ModalRuleTriggerEvent(props) {
    const [event, setEvent] = useState('');
    const [validEvent, setValidEvent] = useState(true);
    const updateEvent = (e) => {
        setEvent(e);
        if (e !== event) {
            setOptions({});
            setValidOptions(true);
            switch (e) {
                case 'Rant':
                    setOptions({ min_amount: 0, max_amount: 0 });
                    break;
                default:
                    setOptions({});
            }
        }
        setValidEvent(true);
    };
    const [options, setOptions] = useState({});
    const [validOptions, setValidOptions] = useState(true);
    const [source, setSource] = useState('');
    const [validSource, setValidSource] = useState(true);
    const updateSource = (s) => {
        setSource(s);
        if (s !== source) {
            setEvent('');
            setValidOptions(true);
        }
        setValidSource(true);
    };
    const [parameters, setParameters] = useState({
        Account: { events: ['Follow'] },
        Channel: { events: ['Follow'] },
        'Live Stream': { events: ['Raid', 'Rant', 'Sub'] },
    });

    useEffect(() => {
        if (props.rule.trigger.on_event === undefined || props.rule.trigger.on_event === null) {
            return;
        }

        const onEvent = props.rule.trigger.on_event;
        switch (true) {
            case onEvent.from_account !== undefined && onEvent.from_account !== null:
                setSource('Account');
                const fromAccount = props.rule.trigger.on_event.from_account;
                switch (true) {
                    case fromAccount.on_follow !== undefined && fromAccount.on_follow !== null:
                        setEvent('Follow');
                        break;
                }
                break;
            case onEvent.from_channel !== undefined && onEvent.from_channel !== null:
                setSource('Channel');
                const fromChannel = props.rule.trigger.on_event.from_channel;
                switch (true) {
                    case fromChannel.on_follow !== undefined && fromChannel.on_follow !== null:
                        setEvent('Follow');
                        break;
                }
                break;
            case onEvent.from_live_stream !== undefined && onEvent.from_live_stream !== null:
                setSource('Live Stream');
                const fromLiveStream = props.rule.trigger.on_event.from_live_stream;
                switch (true) {
                    case fromLiveStream.on_raid !== undefined && fromLiveStream.on_raid !== null:
                        setEvent('Raid');
                        break;
                    case fromLiveStream.on_rant !== undefined && fromLiveStream.on_rant !== null:
                        setEvent('Rant');
                        setOptions(props.rule.trigger.on_event.from_live_stream.on_rant);
                        break;
                    case fromLiveStream.on_sub !== undefined && fromLiveStream.on_sub !== null:
                        setEvent('Sub');
                        break;
                }
                break;
            default:
                return;
        }
    }, []);

    const validRantOptions = () => {
        if (isNaN(options.min_amount) || isNaN(options.max_amount)) {
            setValidOptions(false);
            return false;
        }

        if (options.max_amount !== 0 && options.min_amount > options.max_amount) {
            setValidOptions(false);
            return false;
        }

        return true;
    };

    const fromAccount = () => {
        let from_account = {};
        switch (event) {
            case 'Follow':
                from_account.name = options.page;
                from_account.on_follow = {};
                break;
            default:
                setValidEvent(false);
                return;
        }

        const rule = props.rule;
        if (rule.trigger.on_event == undefined || rule.trigger.on_event == null) {
            rule.trigger.on_event = {};
        }

        rule.trigger.on_event.from_account = from_account;
        rule.trigger.on_event.from_channel = null;
        rule.trigger.on_event.from_live_stream = null;

        props.setRule(rule);
        next('message');
    };

    const fromChannel = () => {
        let from_channel = {};
        switch (event) {
            case 'Follow':
                from_channel.name = options.page;
                from_channel.on_follow = {};
                break;
            default:
                setValidEvent(false);
                return;
        }

        const rule = props.rule;
        if (rule.trigger.on_event == undefined || rule.trigger.on_event == null) {
            rule.trigger.on_event = {};
        }

        rule.trigger.on_event.from_account = null;
        rule.trigger.on_event.from_channel = from_channel;
        rule.trigger.on_event.from_live_stream = null;

        props.setRule(rule);
        next('message');
    };

    const fromLiveStream = () => {
        let from_live_stream = {};
        switch (event) {
            case 'Raid':
                from_live_stream.on_raid = {};
                break;
            case 'Rant':
                if (!validRantOptions()) {
                    return;
                }
                from_live_stream.on_rant = options;
                break;
            case 'Sub':
                from_live_stream.on_sub = {};
                break;
            default:
                setValidEvent(false);
                return;
        }

        const rule = props.rule;
        if (rule.trigger.on_event == undefined || rule.trigger.on_event == null) {
            rule.trigger.on_event = {};
        }

        rule.trigger.on_event.from_account = null;
        rule.trigger.on_event.from_channel = null;
        rule.trigger.on_event.from_live_stream = from_live_stream;

        props.setRule(rule);
        next('message');
    };

    const back = () => {
        props.onBack();
    };

    const next = (stage) => {
        props.setStage(stage);
    };

    const submit = () => {
        switch (source) {
            case 'Account':
                fromAccount();
                break;
            case 'Channel':
                fromChannel();
                break;
            case 'Live Stream':
                fromLiveStream();
                break;
            default:
                setValidSource(false);
        }
    };

    return (
        <Modal
            cancelButton={'Back'}
            onCancel={back}
            onClose={props.onClose}
            show={props.show}
            submitButton={'Next'}
            onSubmit={submit}
            style={{ height: '480px', minHeight: '480px', width: '360px', minWidth: '360px' }}
        >
            <div className='modal-add-account-channel'>
                <span className='modal-add-account-channel-title'>Configure Event</span>
                <div className='chatbot-modal-event-body'>
                    <div className='chatbot-modal-event-body-top'>
                        <div className='chatbot-modal-event-setting'>
                            <label
                                className={
                                    validSource
                                        ? 'chatbot-modal-option-label'
                                        : 'chatbot-modal-option-label-warning'
                                }
                            >
                                Source{!validSource && '*'}
                            </label>
                            <div style={{ width: '250px' }}>
                                <DropDown
                                    options={Object.keys(parameters)}
                                    select={updateSource}
                                    selected={source}
                                />
                            </div>
                        </div>
                        <div className='chatbot-modal-event-setting'>
                            <label
                                className={
                                    validEvent
                                        ? 'chatbot-modal-option-label'
                                        : 'chatbot-modal-option-label-warning'
                                }
                            >
                                Event{!validEvent && '*'}
                            </label>
                            <div style={{ width: '250px' }}>
                                {source !== '' && (
                                    <DropDown
                                        options={parameters[source].events}
                                        select={updateEvent}
                                        selected={event}
                                    />
                                )}
                            </div>
                        </div>
                    </div>
                    <div className='chatbot-modal-event-body-bottom'>
                        <label
                            className={
                                validOptions
                                    ? 'chatbot-modal-event-options-label'
                                    : 'chatbot-modal-event-options-label-warning'
                            }
                        >
                            {validOptions ? 'Options' : 'Verify Options'}
                        </label>
                        <div className='chatbot-modal-event-options'>
                            {event === 'Rant' && (
                                <EventOptionsRant options={options} setOptions={setOptions} />
                            )}
                            {event === 'Follow' && (
                                <EventOptionsFollow
                                    options={options}
                                    setOptions={setOptions}
                                    source={source}
                                />
                            )}
                        </div>
                    </div>
                </div>
                <div></div>
            </div>
        </Modal>
    );
}

function EventOptionsFollow(props) {
    const [accounts, setAccounts] = useState({});
    const [page, setPage] = useState(props.options.page === undefined ? '' : props.options.page);
    const updatePage = (name) => {
        setPage(name);
        props.setOptions({ page: name });
    };

    useEffect(() => {
        AccountList()
            .then((response) => {
                setAccounts(response);
            })
            .catch((error) => {
                setError(error);
            });
    }, []);

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
            if (props.source === 'Account') {
                pages.push(account.account.username);
            }
            if (props.source === 'Channel') {
                const channels = sortChannels(account.channels);
                channels.forEach((channel, j) => {
                    pages.push(channel.name);
                });
            }
        });

        return pages;
    };

    return (
        <div className='modal-add-account-channel-body' style={{ height: '90%' }}>
            <div className='chatbot-modal-pages'>
                {sortPages().map((option, index) => (
                    <div className={'chatbot-modal-page'} key={index}>
                        <button
                            className='chatbot-modal-page-button'
                            onClick={() => updatePage(option)}
                            style={{
                                backgroundColor: page === option ? '#85c742' : '',
                            }}
                        >
                            {option}
                        </button>
                    </div>
                ))}
            </div>
        </div>
    );
}

function EventOptionsRant(props) {
    const [minAmount, setMinAmount] = useState(
        isNaN(props.options.min_amount) ? 0 : props.options.min_amount
    );
    const updateMinAmount = (event) => {
        let amount = parseInt(event.target.value);
        if (isNaN(amount)) {
            amount = 0;
        }

        if (maxAmount !== 0 && amount > maxAmount) {
            setValidMaxAmount(false);
        } else {
            setValidMaxAmount(true);
        }

        setMinAmount(event.target.value);
        props.setOptions({ min_amount: amount, max_amount: maxAmount });
    };
    const [maxAmount, setMaxAmount] = useState(
        isNaN(props.options.max_amount) ? 0 : props.options.max_amount
    );
    const updateMaxAmount = (event) => {
        let amount = parseInt(event.target.value);
        if (isNaN(amount)) {
            amount = 0;
        }

        if (amount !== 0) {
            if (amount < minAmount) {
                setValidMaxAmount(false);
            } else {
                setValidMaxAmount(true);
            }
        } else {
            setValidMaxAmount(true);
        }

        setMaxAmount(amount);
        props.setOptions({ min_amount: minAmount, max_amount: amount });
    };
    const [validMaxAmount, setValidMaxAmount] = useState(true);

    return (
        <>
            <div className='chatbot-modal-setting' style={{ paddingTop: '0px' }}>
                <label className='chatbot-modal-setting-description'>Min rant amount</label>
                <div>
                    <span className='command-rant-amount-symbol'>$</span>
                    <input
                        className='command-rant-amount'
                        onChange={updateMinAmount}
                        placeholder='0'
                        size='4'
                        type='text'
                        value={minAmount === 0 ? '' : minAmount}
                    />
                </div>
            </div>
            <div className='chatbot-modal-setting' style={{ paddingTop: '0px' }}>
                <label
                    className={
                        validMaxAmount
                            ? 'chatbot-modal-setting-description'
                            : 'chatbot-modal-setting-description-warning'
                    }
                >
                    Max rant amount{!validMaxAmount && ' (>= min)'}
                </label>
                <div>
                    <span className='command-rant-amount-symbol'>$</span>
                    <input
                        className='command-rant-amount'
                        onChange={updateMaxAmount}
                        placeholder='0'
                        size='4'
                        type='text'
                        value={maxAmount === 0 ? '' : maxAmount}
                    />
                </div>
            </div>
        </>
    );
}

function ModalRuleTriggerTimer(props) {
    const prependZero = (value) => {
        if (value < 10) {
            return '0' + value;
        }

        return '' + value;
    };

    const intervalToTimer = (interval) => {
        let hours = Math.floor(interval / 3600);
        let minutes = Math.floor(interval / 60 - hours * 60);
        let seconds = Math.floor(interval - hours * 3600 - minutes * 60);

        if (hours !== 0 || minutes !== 0) {
            seconds = prependZero(seconds);
        }
        if (hours !== 0) {
            minutes = prependZero(minutes);
        }
        if (hours === 0) {
            hours = '';
            if (minutes === 0) {
                minutes = '';
                if (seconds === 0) {
                    seconds = '';
                }
            }
        }

        return hours + minutes + seconds;
    };

    const [validTimer, setValidTimer] = useState(true);
    const [timer, setTimer] = useState(
        props.rule.trigger.on_timer !== undefined && props.rule.trigger.on_timer !== null
            ? intervalToTimer(props.rule.trigger.on_timer)
            : ''
    );

    const back = () => {
        // const rule = props.rule;
        // rule.trigger.on_timer = '';
        // props.setRule(rule);
        props.onBack();
    };

    const next = () => {
        if (timer === '') {
            setValidTimer(false);
            return;
        }

        const rule = props.rule;
        rule.trigger.on_timer = timerValToInterval(timer);
        props.setRule(rule);
        props.setStage('message');
    };

    const timerValToInterval = (val) => {
        let prefix = '0'.repeat(6 - val.length);
        let t = prefix + val;

        let hours = parseInt(t.substring(0, 2));
        let minutes = parseInt(t.substring(2, 4));
        let seconds = parseInt(t.substring(4, 6));

        return hours * 3600 + minutes * 60 + seconds;
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
                    <Timer timer={timer} setTimer={setTimer} style={{ fontSize: '24px' }} />
                </div>
                <div style={{ height: '56px' }}></div>
            </div>
        </Modal>
    );
}

function ModalRuleMessage(props) {
    const [error, setError] = useState('');
    const [message, setMessage] = useState(
        props.rule.message !== undefined && props.rule.message !== null ? props.rule.message : {}
    );
    const [refresh, setRefresh] = useState(false);
    const [validFile, setValidFile] = useState(true);
    const [validText, setValidText] = useState(true);

    const back = () => {
        // const rule = props.rule;
        // rule.message = null;
        // props.setRule(rule);
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
        props.setStage('sender');
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
            message.from_text = '';
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
        setRefresh(!refresh);
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

function ModalRuleSender(props) {
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
        // const rule = props.rule;
        // rule.send_as = null;
        // props.setRule(rule);
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
        props.setStage('review');
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
                        <span className='modal-add-account-channel-title'>Choose Sender</span>
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

function ModalRuleReview(props) {
    const [deleteRule, setDeleteRule] = useState(false);
    const [edit, setEdit] = useState(props.edit);
    const updateEdit = (e) => {
        setEdit(e);
        props.setEdit(e);
    };
    const [error, setError] = useState('');

    const back = () => {
        props.onBack();
    };

    const submit = () => {
        props.onSubmit();
    };

    const displayTrigger = () => {
        switch (true) {
            case props.rule.trigger.on_timer !== undefined || props.rule.trigger.on_timer !== null:
                return 'Timer';
            default:
                return 'Error';
        }
    };

    const confirmDelete = () => {
        setDeleteRule(false);
        props.onDelete();
    };

    return (
        <>
            {deleteRule && (
                <SmallModal
                    cancelButton={'Cancel'}
                    onCancel={() => setDeleteRule(false)}
                    onClose={() => setDeleteRule(false)}
                    show={deleteRule}
                    style={{ minWidth: '300px', maxWidth: '200px', maxHeight: '200px' }}
                    title={'Delete Rule'}
                    message={'Are you sure you want to delete this rule?'}
                    submitButton={'OK'}
                    onSubmit={confirmDelete}
                />
            )}
            <Modal
                cancelButton={edit ? 'Back' : ''}
                onCancel={back}
                onClose={props.onClose}
                deleteButton={edit ? '' : 'Delete'}
                deleteActive={true}
                onDelete={() => setDeleteRule(true)}
                show={props.show}
                submitButton={edit ? 'Submit' : 'Edit'}
                onSubmit={edit ? submit : () => updateEdit(true)}
                style={{ height: '480px', minHeight: '480px', width: '360px', minWidth: '360px' }}
            >
                <div className='modal-add-account-channel'>
                    <div className='modal-add-account-channel-header'>
                        <span className='modal-add-account-channel-title'>Review</span>
                    </div>
                    <div className='modal-add-account-channel-body'>
                        <div className='chatbot-modal-review'>
                            <pre>{JSON.stringify(props.rule, null, 2)}</pre>
                        </div>
                    </div>
                    <div style={{ height: '29px' }}></div>
                </div>
            </Modal>
        </>
    );
}

function Command(props) {
    const updateCommand = (e) => {
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

        props.setCommand(command);
    };

    return (
        <input
            className={'command-input'}
            onInput={updateCommand}
            placeholder={'!command'}
            size='8'
            type='text'
            value={props.command}
        />
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
        // let timer = intervalToTimer(props.timer);

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
            style={props.style}
            type='text'
            value={''}
        />
    );
}

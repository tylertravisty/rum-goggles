import { useEffect, useState } from 'react';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import {
    Eye,
    EyeRed,
    EyeSlash,
    Gear,
    Heart,
    Play,
    Pause,
    Star,
    ThumbsDown,
    ThumbsUp,
    ChessRook,
} from '../assets';
import './PageDetails.css';
import {
    ActivateAccount,
    ActivateChannel,
    DeleteAccount,
    DeleteChannel,
    Login,
    Logout,
    UpdateAccountApi,
    UpdateChannelApi,
} from '../../wailsjs/go/main/App';
import { Modal, SmallModal } from './Modal';

function countString(value) {
    switch (true) {
        case value <= 0 || value == undefined:
            return '0';
        case value < 1000:
            return value;
        case value < 1000000:
            return (value / 1000).toFixed(3).slice(0, -2) + 'K';
        case value < 1000000000:
            return (value / 1000000).toFixed(6).slice(0, -5) + 'M';
        default:
            return 'Inf';
    }
}

function PageDetails(props) {
    const [activate, setActivate] = useState(false);
    const [active, setActive] = useState(false);
    const [activity, setActivity] = useState(null);
    const [openApi, setOpenApi] = useState(false);
    const [apiValid, setApiValid] = useState(true);
    const [editingApi, setEditingApi] = useState(false);
    const [editApi, setEditApi] = useState('');
    const updateEditApi = (event) => {
        setEditApi(event.target.value);
    };
    const [openDelete, setOpenDelete] = useState(false);
    const [deleting, setDeleting] = useState(false);
    const [deleteName, setDeleteName] = useState('');
    const updateDeleteName = (event) => {
        if (deleting) {
            return;
        }
        setDeleteName(event.target.value);
    };
    const [details, setDetails] = useState(null);
    const [error, setError] = useState('');
    const [live, setLive] = useState(false);
    const [liveTitle, setLiveTitle] = useState('');
    const [openLogin, setOpenLogin] = useState(false);
    const [loggingIn, setLoggingIn] = useState(false);
    const [loginUsername, setLoginUsername] = useState('');
    const updateLoginUsername = (event) => {
        if (loggingIn) {
            return;
        }
        setLoginUsername(event.target.value);
    };
    const [loginUsernameValid, setLoginUsernameValid] = useState(true);
    const [loginPassword, setLoginPassword] = useState('');
    const updateLoginPassword = (event) => {
        if (loggingIn) {
            return;
        }
        setLoginPassword(event.target.value);
    };
    const [loginPasswordValid, setLoginPasswordValid] = useState(true);
    const [openLogout, setOpenLogout] = useState(false);
    const [loggingOut, setLoggingOut] = useState(false);
    const [settings, setSettings] = useState(false);
    const triggerSettings = () => setSettings(!settings);

    useEffect(() => {
        EventsOn('PageDetails', (event) => {
            setDetails(event);
            // TODO: do I need to reset all editing/logging out/etc. values?
        });

        EventsOn('PageActivity', (event) => {
            setActivity(event);
            if (event !== null) {
                setActive(true);
                if (event.livestreams.length > 0) {
                    setLive(true);
                } else {
                    setLive(false);
                }
            }
        });

        EventsOn('PageActive', (event) => {
            if (event) {
                setActive(true);
            } else {
                setActive(false);
                setActivity(null);
                setLive(false);
            }
        });
    }, []);

    useEffect(() => {
        if (deleting) {
            switch (true) {
                case details.type === 'Channel':
                    DeleteChannel(details.id)
                        .then(() => resetDelete())
                        .catch((error) => {
                            setDeleting(false);
                            setError(error);
                        });
                    return;
                case details.type === 'Account':
                    DeleteAccount(details.id)
                        .then(() => resetDelete())
                        .catch((error) => {
                            setDeleting(false);
                            setError(error);
                        });
                    return;
            }
        }
    }, [deleting]);

    useEffect(() => {
        if (editingApi) {
            switch (true) {
                case details.type === 'Channel':
                    UpdateChannelApi(details.id, editApi)
                        .then(() => resetEditApi())
                        .catch((error) => {
                            setEditingApi(false);
                            setError(error);
                        });
                    return;
                case details.type === 'Account':
                    UpdateAccountApi(details.id, editApi)
                        .then(() => resetEditApi())
                        .catch((error) => {
                            setEditingApi(false);
                            setError(error);
                        });
                    return;
            }
        }
    }, [editingApi]);

    useEffect(() => {
        if (loggingIn && details.type === 'Account') {
            Login(loginUsername, loginPassword)
                .then(() => {
                    resetLogin();
                })
                .catch((error) => {
                    setLoggingIn(false);
                    setError(error);
                });
        } else if (loggingIn && details.type === 'Channel') {
            resetLogin();
        }
    }, [loggingIn]);

    useEffect(() => {
        if (loggingOut && details.type === 'Account') {
            Logout(details.id)
                .catch((error) => {
                    setError(error);
                })
                .finally(() => resetLogout());
        } else if (loggingOut && details.type === 'Channel') {
            resetLogout();
        }
    }, [loggingOut]);

    const activatePage = () => {
        switch (true) {
            case details.type === 'Channel':
                ActivateChannel(details.id).catch((error) => {
                    setError(error);
                });
                return;
            case details.type === 'Account':
                ActivateAccount(details.id).catch((error) => {
                    setError(error);
                });
                return;
        }
    };

    const deletePage = () => {
        if (deleting || details.title !== deleteName) {
            return;
        }

        setDeleting(true);
    };

    const resetDelete = () => {
        setDeleteName('');
        setDeleting(false);
        setOpenDelete(false);
    };

    const submitEditApi = () => {
        if (editingApi) {
            return;
        }

        if (editApi === '') {
            setApiValid(false);
            return;
        }

        setEditingApi(true);
    };

    const closeEditApi = () => {
        if (editingApi) {
            return;
        }

        resetEditApi();
    };

    const resetEditApi = () => {
        setOpenApi(false);
        setApiValid(true);
        setEditApi('');
        setEditingApi(false);
    };

    const login = () => {
        if (loginUsername === '') {
            setLoginUsernameValid(false);
            return;
        }

        if (loginPassword === '') {
            setLoginPasswordValid(false);
            return;
        }

        setLoggingIn(true);
    };

    const closeLogin = () => {
        if (loggingIn) {
            return;
        }

        setOpenLogin(false);
    };

    const resetLogin = () => {
        setLoggingIn(false);
        setOpenLogin(false);
    };

    const logout = () => {
        setLoggingOut(true);
    };

    const closeLogout = () => {
        if (loggingOut) {
            return;
        }

        setOpenLogout(false);
    };

    const resetLogout = () => {
        setLoggingOut(false);
        setOpenLogout(false);
    };

    return (
        <>
            {openLogin && (
                <Modal
                    backgroundClose={true}
                    cancelButton={'Cancel'}
                    onCancel={closeLogin}
                    onClose={closeLogin}
                    show={openLogin}
                    style={{
                        height: '480px',
                        minHeight: '480px',
                        width: '360px',
                        minWidth: '360px',
                    }}
                    submitButton={'Login'}
                    submitLoading={loggingIn}
                    onSubmit={login}
                >
                    <ModalLogin
                        password={loginPassword}
                        passwordValid={loginPasswordValid}
                        updatePassword={updateLoginPassword}
                        username={loginUsername}
                        usernameValid={loginUsernameValid}
                        updateUsername={updateLoginUsername}
                    />
                </Modal>
            )}
            {openLogout && (
                <SmallModal
                    cancelButton={'Cancel'}
                    onCancel={closeLogout}
                    onClose={closeLogout}
                    show={openLogout}
                    style={{ minWidth: '300px', maxWidth: '200px', maxHeight: '200px' }}
                    title={'Logout'}
                    message={'Are you sure you want to log out of ' + details.title + '?'}
                    submitButton={loggingOut ? 'Logging out...' : 'Logout'}
                    onSubmit={logout}
                />
            )}
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
            {openDelete && details !== null && (
                <Modal
                    backgroundClose={true}
                    cancelButton={'Cancel'}
                    onCancel={resetDelete}
                    onClose={resetDelete}
                    deleteButton={deleting ? 'Deleting' : 'Delete'}
                    onDelete={deletePage}
                    deleteActive={details.title === deleteName}
                    pageName={details.title}
                    show={openDelete}
                    style={{
                        height: '350px',
                        minHeight: '350px',
                        width: '350px',
                        minWidth: '350px',
                    }}
                >
                    <div className='modal-delete-page'>
                        <div className='modal-delete-page-header'>
                            <span className='modal-delete-page-title'>Delete page</span>
                            <span className='modal-delete-page-subtitle'>
                                Are you sure you want to delete <b>{details.title}</b>? This cannot
                                be undone. You must type '{details.title}' into the box to delete.
                            </span>
                        </div>
                        <div className='modal-delete-page-body'>
                            <div className='modal-delete-page-input'>
                                <input
                                    className='modal-delete-page-input-text'
                                    onChange={updateDeleteName}
                                    placeholder={details.title}
                                    type={'text'}
                                    value={deleteName}
                                ></input>
                            </div>
                        </div>
                    </div>
                </Modal>
            )}
            {openApi && (
                <Modal
                    backgroundClose={true}
                    cancelButton={'Cancel'}
                    onCancel={closeEditApi}
                    onClose={closeEditApi}
                    show={openApi}
                    style={{
                        height: '480px',
                        minHeight: '480px',
                        width: '360px',
                        minWidth: '360px',
                    }}
                    submitButton={'Submit'}
                    submitLoading={editingApi}
                    onSubmit={submitEditApi}
                >
                    <ModalEditApi
                        apiKey={editApi}
                        updateApiKey={updateEditApi}
                        apiValid={apiValid}
                    />
                </Modal>
            )}
            <div className='page-details'>
                {details !== null && (
                    <>
                        <div className='page-details-header'>
                            <div className='page-details-header-left'>
                                <span className='page-details-header-title'>{details.title}</span>
                                <span className='page-details-header-type'>{details.type}</span>
                            </div>
                            <div className='page-details-header-right'>
                                {details.has_api && (
                                    <button
                                        className='page-details-header-button'
                                        onClick={activatePage}
                                    >
                                        <img
                                            className='page-details-header-icon'
                                            src={active ? Pause : Play}
                                        />
                                    </button>
                                )}
                                <button
                                    className='page-details-header-button'
                                    onClick={triggerSettings}
                                >
                                    <img className='page-details-header-icon' src={Gear} />
                                </button>
                            </div>
                        </div>
                        {settings && (
                            <>
                                <div
                                    className='page-details-settings-background'
                                    onClick={triggerSettings}
                                ></div>
                                <div className='page-details-settings'>
                                    {details.type === 'Account' && (
                                        <button
                                            className='page-details-settings-button'
                                            onClick={() => {
                                                triggerSettings();
                                                if (details.logged_in) {
                                                    setOpenLogout(true);
                                                } else {
                                                    setOpenLogin(true);
                                                }
                                            }}
                                        >
                                            {details.logged_in ? 'Logout' : 'Login'}
                                        </button>
                                    )}
                                    <button
                                        className='page-details-settings-button'
                                        onClick={() => {
                                            triggerSettings();
                                            setOpenApi(true);
                                        }}
                                    >
                                        Edit API key
                                    </button>
                                    <button
                                        className='page-details-settings-button'
                                        onClick={() => {
                                            triggerSettings();
                                            setOpenDelete(true);
                                        }}
                                    >
                                        Delete
                                    </button>
                                </div>
                            </>
                        )}
                        {active && activity !== null && (
                            <>
                                <PageActivity activity={activity} />
                                {live && (
                                    <DetailsFooter
                                        categories={activity.livestreams[0].categories}
                                        dislikes={activity.livestreams[0].dislikes}
                                        likes={activity.livestreams[0].likes}
                                        title={activity.livestreams[0].title}
                                        viewers={activity.livestreams[0].watching_now}
                                    />
                                )}
                            </>
                        )}
                        {!active && (
                            <div className='page-inactive'>
                                <span className='page-inactive-text'>
                                    {details.has_api
                                        ? 'Press play to start API'
                                        : 'Open settings to add API key'}
                                </span>
                            </div>
                        )}
                    </>
                )}
            </div>
        </>
    );
}

export default PageDetails;

function PageActivity(props) {
    const eventDate = (event) => {
        if (event.followed_on) {
            return event.followed_on;
        }
        if (event.subscribed_on) {
            return event.subscribed_on;
        }
    };
    const sortEvents = () => {
        let sorted = [
            ...props.activity.followers.recent_followers,
            ...props.activity.subscribers.recent_subscribers,
        ].sort((a, b) => (eventDate(a) < eventDate(b) ? 1 : -1));

        return sorted;
    };
    return (
        <>
            <div className='page-activity-header'>
                <div className='page-activity-stat'>
                    <span className='page-activity-stat-title'>Followers:</span>
                    <span className='page-activity-stat-text'>
                        {countString(props.activity.followers.num_followers)}
                    </span>
                </div>
                <div className='page-activity-stat'>
                    <span className='page-activity-stat-title'>Subscribers:</span>
                    <span className='page-activity-stat-text'>
                        {countString(props.activity.subscribers.num_subscribers)}
                    </span>
                </div>
            </div>
            <div className='page-activity'>
                <div className='page-activity-list'>
                    {sortEvents().map((event, index) => (
                        <PageEvent event={event} key={index} />
                    ))}
                </div>
            </div>
        </>
    );
}

function PageEvent(props) {
    const dateDate = (date) => {
        const options = { month: 'short' };
        let month = new Intl.DateTimeFormat('en-US', options).format(date);
        let day = date.getDate();
        return month + ' ' + day;
    };

    const dateDay = (date) => {
        let now = new Date();
        let today = now.getDay();
        switch (date.getDay()) {
            case 0:
                return 'Sunday';
            case 1:
                return 'Monday';
            case 2:
                return 'Tuesday';
            case 3:
                return 'Wednesday';
            case 4:
                return 'Thursday';
            case 5:
                return 'Friday';
            case 6:
                return 'Saturday';
        }
    };

    const dateTime = (date) => {
        let now = new Date();
        let today = now.getDay();
        let day = date.getDay();

        if (today !== day) {
            return dateDay(date);
        }

        let hours24 = date.getHours();
        let hours = hours24 % 12 || 12;

        let minutes = date.getMinutes();
        if (minutes < 10) {
            minutes = '0' + minutes;
        }

        let mer = 'pm';
        if (hours24 < 12) {
            mer = 'am';
        }

        return hours + ':' + minutes + ' ' + mer;
    };

    const dateString = (d) => {
        if (isNaN(Date.parse(d))) {
            return 'Who knows?';
        }

        let now = new Date();
        let date = new Date(d);
        // Fix Rumble's timezone problem
        date.setHours(date.getHours() - 4);
        let diff = now - date;
        switch (true) {
            case diff < 0:
                return 'In the future!?';
            case diff < 60000:
                return 'Now';
            case diff < 3600000:
                let minutes = Math.floor(diff / 1000 / 60);
                let postfix = ' mins ago';
                if (minutes == 1) {
                    postfix = ' min ago';
                }
                return minutes + postfix;
            case diff < 86400000:
                return dateTime(date);
            case diff < 604800000:
                return dateDay(date);
            default:
                return dateDate(date);
        }
    };

    return (
        <div className='page-event'>
            <div className='page-event-left'>
                {props.event.followed_on && <img className='page-event-icon' src={Heart}></img>}
                {props.event.subscribed_on && (
                    <img className='page-event-icon' src={ChessRook}></img>
                )}
                <div className='page-event-left-text'>
                    <span className='page-event-username'>{props.event.username}</span>
                    <span className='page-event-description'>
                        {props.event.followed_on && 'Followed you'}
                        {props.event.subscribed_on && 'Subscribed'}
                    </span>
                </div>
            </div>
            <span className='page-event-date'>
                {props.event.followed_on && dateString(props.event.followed_on)}
                {props.event.subscribed_on && dateString(props.event.subscribed_on)}
            </span>
        </div>
    );
}

function DetailsFooter(props) {
    return (
        <div className='page-details-footer'>
            <span className='page-details-footer-title'>{props.title}</span>
            <div className='page-details-footer-stats'>
                <div className='page-details-footer-stat'>
                    <img className='page-details-footer-stat-icon' src={EyeRed} />
                    <span className='page-details-footer-stat-text-red'>
                        {countString(props.viewers)}
                    </span>
                </div>
                <div className='page-details-footer-stat'>
                    <img className='page-details-footer-stat-icon' src={ThumbsUp} />
                    <span className='page-details-footer-stat-text'>
                        {countString(props.likes)}
                    </span>
                </div>
                <div className='page-details-footer-stat'>
                    <img className='page-details-footer-stat-icon' src={ThumbsDown} />
                    <span className='page-details-footer-stat-text'>
                        {countString(props.dislikes)}
                    </span>
                </div>
            </div>
            <div className='page-details-footer-categories'>
                <span className='page-details-footer-category'>
                    {props.categories.primary.title}
                </span>
                <span className='page-details-footer-category'>
                    {props.categories.secondary.title}
                </span>
            </div>
        </div>
    );
}

function ModalEditApi(props) {
    const [showKey, setShowKey] = useState(false);
    const updateShowKey = () => setShowKey(!showKey);

    return (
        <div className='modal-add-account-channel'>
            <div className='modal-add-account-channel-header'>
                <span className='modal-add-account-channel-title'>Edit API Key</span>
                <span className='modal-add-account-channel-subtitle'>Enter new API key below</span>
            </div>
            <div className='modal-add-account-channel-body'>
                {props.apiValid === false ? (
                    <label className='modal-add-channel-label-warning'>
                        API KEY - Please enter a valid API key
                    </label>
                ) : (
                    <label className='modal-add-channel-label'>API KEY</label>
                )}
                <div className='modal-add-channel-key'>
                    <input
                        className='modal-add-channel-key-input'
                        onChange={props.updateApiKey}
                        placeholder={'Enter API key'}
                        type={showKey ? 'text' : 'password'}
                        value={props.apiKey}
                    ></input>
                    <button className='modal-add-channel-key-show' onClick={updateShowKey}>
                        <img
                            className='modal-add-channel-key-show-icon'
                            src={showKey ? EyeSlash : Eye}
                        />
                    </button>
                </div>
                <span className='modal-add-channel-description'>API KEYS SHOULD LOOK LIKE</span>
                <span className='modal-add-channel-description-subtext'>
                    https://rumble.com/-livestream-api/get-data?key=really-long_string-of_random-characters
                </span>
            </div>
            <div></div>
        </div>
    );
}

function ModalLogin(props) {
    const [showPassword, setShowPassword] = useState(false);
    const updateShowPassword = () => setShowPassword(!showPassword);

    return (
        <div className='modal-add-account-channel'>
            <div className='modal-add-account-channel-header'>
                <span className='modal-add-account-channel-title'>Login</span>
                <span className='modal-add-account-channel-subtitle'>
                    Log into your Rumble account
                </span>
            </div>
            <div className='modal-add-account-channel-body'>
                {props.usernameValid === false ? (
                    <label className='modal-add-account-channel-label-warning'>
                        USERNAME - Please enter a valid username
                    </label>
                ) : (
                    <label className='modal-add-account-channel-label'>USERNAME</label>
                )}
                <div className='modal-add-account-channel-input'>
                    <input
                        className='modal-add-account-channel-input-text'
                        onChange={!props.loading && props.updateUsername}
                        placeholder={'Username'}
                        type={'text'}
                        value={props.username}
                    ></input>
                </div>
                {props.passwordValid === false ? (
                    <label className='modal-add-account-channel-label-warning'>
                        PASSWORD - Please enter a valid password
                    </label>
                ) : (
                    <label className='modal-add-account-channel-label'>PASSWORD</label>
                )}
                <div className='modal-add-account-channel-input'>
                    <input
                        className='modal-add-account-channel-input-password'
                        onChange={!props.loading && props.updatePassword}
                        placeholder={'Password'}
                        type={showPassword ? 'text' : 'password'}
                        value={props.password}
                    ></input>
                    <button
                        className='modal-add-account-channel-input-show'
                        onClick={updateShowPassword}
                    >
                        <img
                            className='modal-add-account-channel-input-show-icon'
                            src={showPassword ? EyeSlash : Eye}
                        />
                    </button>
                </div>
            </div>
            <div></div>
        </div>
    );
}

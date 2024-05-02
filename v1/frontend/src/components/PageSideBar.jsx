import { useEffect, useState } from 'react';
import { Modal, SmallModal } from './Modal';
import {
    AccountList,
    AddPage,
    Login,
    OpenAccount,
    OpenChannel,
    PageStatus,
} from '../../wailsjs/go/main/App';
import { EventsOff, EventsOn } from '../../wailsjs/runtime/runtime';

import {
    ChevronRight,
    CircleGreenBackground,
    CircleRedBackground,
    Eye,
    EyeSlash,
    PlusCircle,
} from '../assets';
import './PageSideBar.css';

function PageSideBar(props) {
    const [accounts, setAccounts] = useState({});
    const [error, setError] = useState('');
    const [addOpen, setAddOpen] = useState(false);
    // const [refresh, setRefresh] = useState(false);
    const [scrollY, setScrollY] = useState(0);

    useEffect(() => {
        EventsOn('PageSideBarAccounts', (event) => {
            setAccounts(event);
        });
    }, []);

    useEffect(() => {
        AccountList()
            .then((response) => {
                setAccounts(response);
            })
            .catch((error) => {
                setError(error);
            });
    }, []);

    const sortAccounts = () => {
        let keys = Object.keys(accounts);

        let sorted = [...keys].sort((a, b) =>
            accounts[a].account.username.toLowerCase() > accounts[b].account.username.toLowerCase()
                ? 1
                : -1
        );

        return sorted;
    };

    const handleScroll = (event) => {
        setScrollY(event.target.scrollTop);
    };

    const openAccount = (account) => {
        OpenAccount(account.id).catch((error) => setError(error));
    };

    const openChannel = (channel) => {
        OpenChannel(channel.id).catch((error) => setError(error));
    };

    return (
        <>
            {addOpen && (
                <ModalAdd
                    onClose={() => setAddOpen(false)}
                    onRefresh={() => {
                        setRefresh(!refresh);
                    }}
                    show={addOpen}
                />
            )}
            <div className='page-sidebar'>
                <div className='page-sidebar-body' onScroll={handleScroll}>
                    {sortAccounts().map((account, index) => (
                        <AccountChannels
                            account={accounts[account]}
                            key={index}
                            openAccount={openAccount}
                            openChannel={openChannel}
                            scrollY={scrollY}
                            top={index === 0}
                        />
                    ))}
                </div>
                <div className='page-sidebar-footer'>
                    <ButtonIcon
                        hoverText={'Add an account/channel'}
                        onClick={() => setAddOpen(true)}
                        scrollY={0}
                    />
                </div>
            </div>
        </>
    );
}

export default PageSideBar;

function AccountChannels(props) {
    const sortChannels = () => {
        let sorted = [...props.account.channels].sort((a, b) =>
            a.name.toLowerCase() > b.name.toLowerCase() ? 1 : -1
        );

        return sorted;
    };

    if (props.account.account !== undefined) {
        return (
            <div
                className='page-sidebar-account-list'
                style={props.top ? { borderTop: 'none' } : {}}
            >
                <button
                    className='page-sidebar-button'
                    key={0}
                    onClick={() => props.openAccount(props.account.account)}
                >
                    <AccountIcon account={props.account.account} scrollY={props.scrollY} />
                </button>
                {sortChannels().map((channel, index) => (
                    <button
                        className='page-sidebar-button'
                        key={index + 1}
                        onClick={() => props.openChannel(channel)}
                    >
                        <ChannelIcon channel={channel} scrollY={props.scrollY} />
                    </button>
                ))}
            </div>
        );
    }
}

function AccountIcon(props) {
    const [apiActive, setApiActive] = useState(false);
    const [hover, setHover] = useState(false);
    const [isLive, setIsLive] = useState(false);
    const [loggedIn, setLoggedIn] = useState(props.account.cookies !== null);
    const [username, setUsername] = useState(props.account.username);

    const iconBorder = () => {
        if (!apiActive) {
            return '3px solid #3377cc';
        }
        if (isLive) {
            return '3px solid #85c742';
        } else {
            return '3px solid #f23160';
        }
    };

    const pageName = (name) => {
        if (name === undefined) return;
        return '/user/' + name;
    };

    useEffect(() => {
        if (username !== props.account.username) {
            EventsOff(
                'ApiActive-' + pageName(username),
                'LoggedIn-' + pageName(username),
                'PageLive-' + pageName(username)
            );
            setApiActive(false);
            setIsLive(false);
        }

        EventsOn('ApiActive-' + pageName(props.account.username), (event) => {
            setApiActive(event);
        });

        EventsOn('LoggedIn-' + pageName(props.account.username), (event) => {
            setLoggedIn(event);
        });

        EventsOn('PageLive-' + pageName(props.account.username), (event) => {
            setIsLive(event);
        });

        setUsername(props.account.username);
    }, [props.account.username]);

    useEffect(() => {
        if (username !== '') {
            PageStatus(pageName(username));
        }
    }, [username]);

    return (
        <div
            className='page-sidebar-icon'
            onMouseEnter={() => setHover(true)}
            onMouseLeave={() => setHover(false)}
        >
            {props.account.profile_image === null ? (
                <span className='page-sidebar-icon-initial' style={{ border: iconBorder() }}>
                    {props.account.username[0].toUpperCase()}
                </span>
            ) : (
                <img
                    className='page-sidebar-icon-image'
                    src={props.account.profile_image}
                    style={{ border: iconBorder() }}
                />
            )}
            <img
                className='page-sidebar-icon-account'
                src={loggedIn ? CircleGreenBackground : CircleRedBackground}
            />
            {hover && <HoverName name={pageName(username)} scrollY={props.scrollY} />}
        </div>
    );
}

function ButtonIcon(props) {
    const [hover, setHover] = useState(false);

    return (
        <div
            className='page-sidebar-icon'
            onMouseEnter={() => setHover(true)}
            onMouseLeave={() => setHover(false)}
        >
            <button className='page-sidebar-button' onClick={props.onClick}>
                <img className='page-sidebar-button-icon' src={PlusCircle} />
            </button>
            {hover && <HoverName name={props.hoverText} scrollY={props.scrollY} />}
        </div>
    );
}

function ChannelIcon(props) {
    const [apiActive, setApiActive] = useState(false);
    const [channelName, setChannelName] = useState(props.channel.name);
    const [hover, setHover] = useState(false);
    const [isLive, setIsLive] = useState(false);

    const iconBorder = () => {
        if (!apiActive) {
            return '3px solid #3377cc';
        }
        if (isLive) {
            return '3px solid #85c742';
        } else {
            return '3px solid #f23160';
        }
    };

    const pageName = (name) => {
        if (name === undefined) return;
        return '/c/' + name.replace(/\s/g, '');
    };

    useEffect(() => {
        if (channelName !== props.channel.name) {
            EventsOff('PageLive-' + pageName(channelName), 'ApiActive-' + pageName(channelName));
            setApiActive(false);
            setIsLive(false);
        }

        EventsOn('PageLive-' + pageName(props.channel.name), (event) => {
            setIsLive(event);
        });

        EventsOn('ApiActive-' + pageName(props.channel.name), (event) => {
            setApiActive(event);
        });

        setChannelName(props.channel.name);
    }, [props.channel.name]);

    useEffect(() => {
        if (channelName !== '') {
            PageStatus(pageName(channelName));
        }
    }, [channelName]);

    return (
        <div
            className='page-sidebar-icon'
            onMouseEnter={() => setHover(true)}
            onMouseLeave={() => setHover(false)}
        >
            {props.channel.profile_image === null ? (
                <span className='page-sidebar-icon-initial' style={{ border: iconBorder() }}>
                    {props.channel.name[0].toUpperCase()}
                </span>
            ) : (
                <img
                    className='page-sidebar-icon-image'
                    src={props.channel.profile_image}
                    style={{ border: iconBorder() }}
                />
            )}
            {hover && <HoverName name={pageName(channelName)} scrollY={props.scrollY} />}
        </div>
    );
}

function HoverName(props) {
    return (
        <div
            className='page-sidebar-icon-hover'
            style={{ transform: 'translate(75px, -' + (50 + props.scrollY) + 'px)' }}
        >
            <span className='page-sidebar-icon-hover-text'>{props.name}</span>
        </div>
    );
}

function ModalAdd(props) {
    const [accountPassword, setAccountPassword] = useState('');
    const [accountPasswordValid, setAccountPasswordValid] = useState(true);
    const updateAccountPassword = (event) => {
        if (loading()) {
            return;
        }
        setAccountPassword(event.target.value);
    };
    const [accountUsername, setAccountUsername] = useState('');
    const [accountUsernameValid, setAccountUsernameValid] = useState(true);
    const updateAccountUsername = (event) => {
        if (loading()) {
            return;
        }
        setAccountUsername(event.target.value);
    };
    const [addAccountLoading, setAddAccountLoading] = useState(false);
    const [addChannelLoading, setAddChannelLoading] = useState(false);
    const [channelKey, setChannelKey] = useState('');
    const [channelKeyValid, setChannelKeyValid] = useState(true);
    const updateChannelKey = (event) => {
        if (loading()) {
            return;
        }
        setChannelKey(event.target.value);
    };
    const [error, setError] = useState('');
    const [stage, setStage] = useState('start');

    useEffect(() => {
        if (addAccountLoading) {
            Login(accountUsername, accountPassword)
                .then(() => {
                    reset();
                    props.onClose();
                    //props.onRefresh();
                })
                .catch((error) => {
                    setAddAccountLoading(false);
                    setError(error);
                });
        }
    }, [addAccountLoading]);

    useEffect(() => {
        if (addChannelLoading) {
            AddPage(channelKey)
                .then(() => {
                    reset();
                    props.onClose();
                    //props.onRefresh();
                })
                .catch((error) => {
                    setAddChannelLoading(false);
                    setError(error);
                });
        }
    }, [addChannelLoading]);

    const back = () => {
        if (loading()) {
            return;
        }
        reset();
    };

    const close = () => {
        if (loading()) {
            return;
        }
        reset();
        props.onClose();
    };

    const reset = () => {
        setStage('start');
        resetAccount();
        resetChannel();
    };

    const add = () => {
        switch (stage) {
            case 'account':
                addAccount();
                break;
            case 'channel':
                addChannel();
                break;
            default:
                close();
        }
    };

    const addAccount = () => {
        if (loading()) {
            return;
        }

        if (accountUsername === '') {
            setAccountUsernameValid(false);
            return;
        }

        if (accountPassword === '') {
            setAccountPasswordValid(false);
            return;
        }

        setAddAccountLoading(true);
    };

    const addChannel = () => {
        if (loading()) {
            return;
        }

        if (channelKey === '') {
            setChannelKeyValid(false);
            return;
        }

        setAddChannelLoading(true);
    };

    const loading = () => {
        return addAccountLoading || addChannelLoading;
    };

    const resetAccount = () => {
        setAccountPassword('');
        setAccountPasswordValid(true);
        setAccountUsername('');
        setAccountUsernameValid(true);
        setAddAccountLoading(false);
    };

    const resetChannel = () => {
        setChannelKey('');
        setChannelKeyValid(true);
        setAddChannelLoading(false);
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
                cancelButton={stage !== 'start' ? 'Back' : ''}
                onCancel={back}
                onClose={close}
                show={props.show}
                style={{ height: '480px', minHeight: '480px', width: '360px', minWidth: '360px' }}
                submitButton={stage !== 'start' ? 'Add' : ''}
                submitLoading={loading()}
                onSubmit={add}
            >
                {stage === 'start' && <ModalAddStart setStage={setStage} />}
                {stage === 'account' && (
                    <ModalAddAccount
                        accountPassword={accountPassword}
                        accountPasswordValid={accountPasswordValid}
                        updateAccountPassword={updateAccountPassword}
                        accountUsername={accountUsername}
                        accountUsernameValid={accountUsernameValid}
                        updateAccountUsername={updateAccountUsername}
                    />
                )}
                {stage === 'channel' && (
                    <ModalAddChannel
                        channelKey={channelKey}
                        channelKeyValid={channelKeyValid}
                        updateChannelKey={updateChannelKey}
                    />
                )}
            </Modal>
        </>
    );
}

function ModalAddAccount(props) {
    const [showKey, setShowKey] = useState(false);
    const updateShowKey = () => setShowKey(!showKey);
    const [showPassword, setShowPassword] = useState(false);
    const updateShowPassword = () => setShowPassword(!showPassword);

    return (
        <div className='modal-add-account-channel'>
            <div className='modal-add-account-channel-header'>
                <span className='modal-add-account-channel-title'>Add Account</span>
                <span className='modal-add-account-channel-subtitle'>
                    Log into your Rumble account
                </span>
            </div>
            <div className='modal-add-account-channel-body'>
                {props.accountUsernameValid === false ? (
                    <label className='modal-add-account-channel-label-warning'>
                        USERNAME - Please enter a valid username
                    </label>
                ) : (
                    <label className='modal-add-account-channel-label'>USERNAME</label>
                )}
                <div className='modal-add-account-channel-input'>
                    <input
                        className='modal-add-account-channel-input-text'
                        onChange={!props.loading && props.updateAccountUsername}
                        placeholder={'Username'}
                        type={'text'}
                        value={props.accountUsername}
                    ></input>
                </div>
                {props.accountPasswordValid === false ? (
                    <label className='modal-add-account-channel-label-warning'>
                        PASSWORD - Please enter a valid password
                    </label>
                ) : (
                    <label className='modal-add-account-channel-label'>PASSWORD</label>
                )}
                <div className='modal-add-account-channel-input'>
                    <input
                        className='modal-add-account-channel-input-password'
                        onChange={!props.loading && props.updateAccountPassword}
                        placeholder={'Password'}
                        type={showPassword ? 'text' : 'password'}
                        value={props.accountPassword}
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

function ModalAddChannel(props) {
    const [showKey, setShowKey] = useState(false);
    const updateShowKey = () => setShowKey(!showKey);

    return (
        <div className='modal-add-account-channel'>
            <div className='modal-add-account-channel-header'>
                <span className='modal-add-account-channel-title'>Add Channel</span>
                <span className='modal-add-account-channel-subtitle'>
                    Copy an API key below to add a channel
                </span>
            </div>
            <div className='modal-add-account-channel-body'>
                {props.channelKeyValid === false ? (
                    <label className='modal-add-channel-label-warning'>
                        API KEY - Please enter a valid API key
                    </label>
                ) : (
                    <label className='modal-add-channel-label'>API KEY</label>
                )}
                <div className='modal-add-channel-key'>
                    <input
                        className='modal-add-channel-key-input'
                        onChange={!props.loading && props.updateChannelKey}
                        placeholder={'Enter API key'}
                        type={showKey ? 'text' : 'password'}
                        value={props.channelKey}
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

function ModalAddStart(props) {
    return (
        <div className='modal-add-account-channel'>
            <span className='modal-add-account-channel-title'>Add an Account or Channel</span>
            <div className='modal-add-account-channel-body'>
                <button
                    className='modal-add-account-channel-button'
                    onClick={() => props.setStage('account')}
                >
                    <div className='modal-add-account-channel-button-left'>
                        <span>Add Account</span>
                    </div>
                    <img
                        className='modal-add-account-channel-button-right-icon'
                        src={ChevronRight}
                    />
                </button>
                <button
                    className='modal-add-account-channel-button'
                    onClick={() => props.setStage('channel')}
                >
                    <div className='modal-add-account-channel-button-left'>
                        <span>Add Channel</span>
                    </div>
                    <img
                        className='modal-add-account-channel-button-right-icon'
                        src={ChevronRight}
                    />
                </button>
            </div>
            <div></div>
        </div>
    );
}

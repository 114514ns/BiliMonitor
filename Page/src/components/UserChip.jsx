import React, {useEffect, useRef} from 'react';
import {Avatar, Chip, Dropdown, DropdownItem, DropdownMenu, DropdownTrigger, Tooltip} from "@heroui/react";
import {CheckIcon} from "../pages/ChatPage";

function UserChip(props) {
    return (
        <div className={'mt-2 w-full'}>
                <div >
                    <p className={'font-medium w-full'} >{props.props.FromName}</p>
                    {(
                        <div className={'flex flex-row align-middle'}>
                            <Avatar
                                src={`${AVATAR_API}${props.props.FromId}`}
                                onClick={() => {
                                    toSpace(props.props.FromId);
                                }}/>

                            {props.props.MedalLevel ?              <Chip
                                startContent={props.props.MedalLevel != 0 ?<img src={getGuardIcon(props.props.GuardLevel)}/>:<CheckIcon size={18}/> }
                                variant="faded"
                                onClick={() => {
                                    toSpace(props.props.LiverID);
                                }}
                                style={{background: getColor(props.props.MedalLevel), color: 'white', marginLeft: '8px',marginTop:'4px'}}
                            >
                                {props.props.MedalName}
                                <span className="ml-2 text-xs font-bold px-2 py-0.5 rounded-full">
                                                            {props.props.MedalLevel}
                                                        </span>
                            </Chip>:<></>}
                        </div>
                    )}
                </div>

        </div>
    );
}

export default UserChip;
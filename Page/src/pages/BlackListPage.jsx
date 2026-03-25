import React, {useEffect} from 'react';
import {addToast, Autocomplete, AutocompleteItem, Avatar, Button, Switch, ToastProvider} from "@heroui/react";
import axios from "axios";

function BlackListPage(props) {
    const [upper,setUpper] = React.useState('');
    const [user,setUser] = React.useState('');
    const [suggestion,setSuggestion] = React.useState([]);
    const [status,setStatus] = React.useState(false);

    return (
        <div>
            <ToastProvider placement={'top-right'}/>
            {[0,1].map((x,index) => (
                <div key={index}>
                    <Autocomplete onInputChange={e => {
                        if (!e.includes("'")) {
                            axios.get('/api/search/user?key=' + e).then((res) => {
                                
                                setSuggestion(prevState => {
                                    const newState = [...prevState];
                                    newState[index] = res.data.data;
                                    return newState;
                                });
                            })
                        }

                    }} className={'max-w-xs mt-2'} onSelectionChange={e => {
                        if (index === 0) {
                            setUpper(e)
                        } else {
                            setUser(e)
                        }
                    }}>
                        {(suggestion[index]??[]).map(item => {
                            return (
                                <AutocompleteItem textValue={item.UName} key={item.UID}>
                                    <div className={'flex flex-row'}>
                                        <Avatar src={`${AVATAR_API}${item.UID}`} alt={item.UName} />
                                        <div className={'flex flex-col ml-2'}>
                                            <p>{item.UName}</p>
                                           <p> {item.Fans.toLocaleString()} Fans</p>
                                        </div>
                                    </div>
                                </AutocompleteItem>
                            )
                        })}

                    </Autocomplete>

                </div>

            ))}
            <Button  onClick={() => {
                axios.get(`/api/black?up=${upper}&user=${user}`).then((res) => {
                    setStatus(res.data.result)
                    addToast({
                        title: "Toast title",
                        description:  res.data.result?'杀！':'状态正常',
                        color: res.data.result?'danger':'success',
                    })
                })
            }} className={'mt-2'}>
                Check Relation
            </Button>

        </div>
    );
}

export default BlackListPage;
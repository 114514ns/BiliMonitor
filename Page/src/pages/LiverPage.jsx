import React from 'react';
import {useParams} from "react-router-dom";

function LiverPage(props) {
    let {id} = useParams();
    return (
        <div>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-2 text-sm">
                <div
                    className="rounded-xl bg-gray-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">开始时间<br/><span
                    className="font-semibold">{1}</span>
                </div>
                <div
                    className="rounded-xl bg-gray-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">结束时间<br/>
                    <span className="font-semibold">{2}</span>
                </div>
                <div
                    className="rounded-xl bg-gray-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">时长<br/>
                    <span
                        className="font-semibold">{3}</span>
                </div>
            </div>
        </div>
    );
}

export default LiverPage;
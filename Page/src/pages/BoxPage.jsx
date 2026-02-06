import React from 'react';
import {DatePicker} from "@heroui/react";
import {parseDate} from "@internationalized/date";

function BoxPage(props) {

    const [value, setValue] = React.useState(parseDate(`${new Date().getFullYear()}-${(new Date().getMonth() + 1 + '').padStart(2,0)}-${(new Date().getDate() + '').padStart(2,0)}`));
    return (
        <div>
            <DatePicker
                className="max-w-[284px]"
                value={value}
                onChange={(e => {
                    setValue(e);
                })}
                label="Date"
            />
        </div>
    );
}

export default BoxPage;
CREATE FUNCTION myrpad(data varchar2,num IN integer,c varchar2) 
RETURN varchar2 
IS acc_bal NUMBER(11,2);
BEGIN 
	if data is null then
   		return Repeat(c,num);
	else
		return data||repeat(c,Length(data)-num);
	end if;
END;
/
